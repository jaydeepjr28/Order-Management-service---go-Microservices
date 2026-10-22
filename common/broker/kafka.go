package broker

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

const (
	MaxRetryCount      = 3
	KafkaMaxRetryCount = MaxRetryCount
	DLQTopic           = "dlq.main"
	RetryCountHeader   = "x-retry-count"
)

// Producer wraps kafka.Writer for publishing events
type Producer struct {
	writer *kafka.Writer
}

// NewProducer initializes a thread-safe Kafka writer
func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
	return &Producer{writer: writer}
}

// Publish publishes a message with OpenTelemetry trace context injection
func (p *Producer) Publish(ctx context.Context, topic string, key []byte, payload []byte) error {
	headers := InjectKafkaHeaders(ctx)
	msg := kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   payload,
		Headers: headers,
		Time:    time.Now(),
	}
	return p.writer.WriteMessages(ctx, msg)
}

// PublishWithHeaders publishes a message with custom headers (e.g. for retries / DLQ)
func (p *Producer) PublishWithHeaders(ctx context.Context, topic string, key []byte, payload []byte, headers []kafka.Header) error {
	msg := kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   payload,
		Headers: headers,
		Time:    time.Now(),
	}
	return p.writer.WriteMessages(ctx, msg)
}

// Close gracefully closes the Kafka producer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// NewConsumerGroup creates a reader for a consumer group listening to multiple topics
func NewConsumerGroup(brokers []string, groupID string, topics []string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                brokers,
		GroupID:                groupID,
		GroupTopics:            topics,
		MinBytes:               1,
		MaxBytes:               10e6, // 10MB
		CommitInterval:         time.Second,
		StartOffset:            kafka.FirstOffset,
		WatchPartitionChanges:  true,
		PartitionWatchInterval: 5 * time.Second,
	})
}

// NewSingleTopicConsumer creates a reader for a consumer group on a single topic
func NewSingleTopicConsumer(brokers []string, groupID string, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                brokers,
		GroupID:                groupID,
		Topic:                  topic,
		MinBytes:               1,
		MaxBytes:               10e6,
		CommitInterval:         time.Second,
		StartOffset:            kafka.FirstOffset,
		WatchPartitionChanges:  true,
		PartitionWatchInterval: 5 * time.Second,
	})
}

// EnsureTopics ensures that the specified topics exist on the Kafka broker
func EnsureTopics(brokers []string, topics ...string) error {
	if len(brokers) == 0 {
		return nil
	}
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, t := range topics {
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     3,
			ReplicationFactor: 1,
		}
	}

	return controllerConn.CreateTopics(topicConfigs...)
}

// HandleKafkaRetry handles exponential backoff retry and DLQ routing
func HandleKafkaRetry(ctx context.Context, producer *Producer, originalTopic string, msg kafka.Message) error {
	headers := msg.Headers
	var retryCount int64 = 0
	foundIndex := -1

	for i, h := range headers {
		if h.Key == RetryCountHeader {
			var err error
			retryCount, err = strconv.ParseInt(string(h.Value), 10, 64)
			if err != nil {
				retryCount = 0
			}
			foundIndex = i
			break
		}
	}

	retryCount++
	newHeader := kafka.Header{
		Key:   RetryCountHeader,
		Value: []byte(strconv.FormatInt(retryCount, 10)),
	}

	if foundIndex >= 0 {
		headers[foundIndex] = newHeader
	} else {
		headers = append(headers, newHeader)
	}

	log.Printf("[Kafka] Retrying message on topic %s, retry count: %d", originalTopic, retryCount)

	if retryCount >= KafkaMaxRetryCount {
		log.Printf("[Kafka] Max retry count (%d) reached. Moving message to DLQ topic: %s", KafkaMaxRetryCount, DLQTopic)
		return producer.PublishWithHeaders(ctx, DLQTopic, msg.Key, msg.Value, headers)
	}

	time.Sleep(time.Duration(retryCount) * time.Second)
	return producer.PublishWithHeaders(ctx, originalTopic, msg.Key, msg.Value, headers)
}

// --- OpenTelemetry Tracing Header Carrier for Kafka ---

type KafkaHeaderCarrier []kafka.Header

func (c *KafkaHeaderCarrier) Get(key string) string {
	for _, h := range *c {
		if strings.EqualFold(h.Key, key) {
			return string(h.Value)
		}
	}
	return ""
}

func (c *KafkaHeaderCarrier) Set(key string, val string) {
	for i, h := range *c {
		if strings.EqualFold(h.Key, key) {
			(*c)[i].Value = []byte(val)
			return
		}
	}
	*c = append(*c, kafka.Header{Key: key, Value: []byte(val)})
}

func (c *KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(*c))
	for _, h := range *c {
		keys = append(keys, h.Key)
	}
	return keys
}

// InjectKafkaHeaders injects trace context into Kafka message headers
func InjectKafkaHeaders(ctx context.Context) []kafka.Header {
	carrier := KafkaHeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &carrier)
	return []kafka.Header(carrier)
}

// ExtractKafkaHeaders extracts trace context from Kafka message headers
func ExtractKafkaHeaders(ctx context.Context, headers []kafka.Header) context.Context {
	carrier := KafkaHeaderCarrier(headers)
	return otel.GetTextMapPropagator().Extract(ctx, &carrier)
}
