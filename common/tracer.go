package common

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// SetGlobalTracer configures OpenTelemetry for the microservice.
// It sets up an OTLP trace exporter over HTTP, which sends traces to Jaeger.
// By initializing this globally, all spans created using otel.Tracer(name)
// will be exported automatically, enabling distributed tracing across the system.
func SetGlobalTracer(ctx context.Context, serviceName, exporterEndpoint string) error {
	// Create an insecure HTTP client to send trace data to the telemetry backend
	client := otlptracehttp.NewClient(
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(exporterEndpoint))

	// Initialize the OTLP trace exporter
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return err
	}

	// Create a TracerProvider that batches trace exports
	// and tags all traces with the specific microservice name.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// Register our TracerProvider as the global provider so any library
	// can access it via otel.GetTracerProvider().
	otel.SetTracerProvider(tp)
	
	// Set the global propagator to TraceContext to ensure trace IDs are
	// passed along in HTTP headers when calling other services.
	otel.SetTextMapPropagator(propagation.TraceContext{})
	
	return nil
}
