package broker

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

const MaxRetryCount = 3
const DLQ = "dlq_main"

// Connect establishes a connection to RabbitMQ, opens a channel, and declares the 
// required exchanges (OrderCreatedEvent, OrderPaidEvent) as well as the Dead Letter Queue.
// It returns the channel and a function to close the connection gracefully.
func Connect(user, pass, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, pass, host, port)

	conn, err := amqp.Dial(address)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	// Declare OrderCreatedEvent as a fanout exchange so multiple services (payments, notification) can listen to it.
	err = ch.ExchangeDeclare(OrderCreatedEvent, "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(OrderPaidEvent, "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(UserRegisteredEvent, "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = createDLQAndDLX(ch)
	if err != nil {
		log.Fatal(err)
	}

	return ch, conn.Close
}

// HandleRetry is called when processing a message fails. It increments the "x-retry-count"
// header on the message. If the retry count exceeds MaxRetryCount, the message is routed to the Dead Letter Queue (DLQ).
// Otherwise, it waits (with exponential backoff) and republishes the message to the original exchange.
func HandleRetry(ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers == nil {
		d.Headers = amqp.Table{}
	}

	retryCount, ok := d.Headers["x-retry-count"].(int64)
	if !ok {
		retryCount = 0
	}
	retryCount++
	d.Headers["x-retry-count"] = retryCount

	log.Printf("Retrying message %s, retry count: %d", d.Body, retryCount)

	if retryCount >= MaxRetryCount {
		log.Printf("Moving message to DLQ %s", DLQ)

		return ch.PublishWithContext(context.Background(), "", DLQ, false, false, amqp.Publishing{
			ContentType:  "application/json",
			Headers:      d.Headers,
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})
	}

	time.Sleep(time.Second * time.Duration(retryCount))

	return ch.PublishWithContext(
		context.Background(),
		d.Exchange,
		d.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Headers:      d.Headers,
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

// createDLQAndDLX creates a Dead Letter Exchange (DLX) and a Dead Letter Queue (DLQ),
// and binds the DLQ to the DLX. Unprocessable messages are sent here to prevent message loss.
func createDLQAndDLX(ch *amqp.Channel) error {
	dlx := "dlx_main"
	err := ch.ExchangeDeclare(
		dlx,     
		"fanout",
		true,    
		false,   
		false,   
		false,   
		nil,     
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		DLQ,  
		true, 
		false,
		false,
		false,
		nil,  
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		DLQ,   
		"",    
		dlx,   
		false,
		nil,
	)
	return err
}

// AmqpHeaderCarrier implements OpenTelemetry's TextMapCarrier interface for AMQP headers.
// This allows trace IDs (like Jaeger tracing context) to be passed across RabbitMQ message boundaries.
type AmqpHeaderCarrier map[string]interface{}

func (a AmqpHeaderCarrier) Get(k string) string {
	value, ok := a[k]
	if !ok {
		return ""
	}

	return value.(string)
}

func (a AmqpHeaderCarrier) Set(k string, v string) {
	a[k] = v
}

func (a AmqpHeaderCarrier) Keys() []string {
	keys := make([]string, len(a))
	i := 0

	for k := range a {
		keys[i] = k
		i++
	}

	return keys
}

// InjectAMQPHeaders extracts the trace context from the Go context and injects it into a map
// of AMQP headers. This should be called right before publishing a message.
func InjectAMQPHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(AmqpHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

// ExtractAMQPHeader takes the headers from a consumed AMQP message and extracts the trace context,
// wrapping it into a new Go context. This should be called immediately after consuming a message.
func ExtractAMQPHeader(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, AmqpHeaderCarrier(headers))
}
