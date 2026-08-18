package discovery

import (
	"context"
	"log"
	"math/rand"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceConnection connects to a specific microservice via gRPC.
// It leverages the provided Registry to discover active instances of the service,
// and randomly selects one instance to implement basic client-side load balancing.
func ServiceConnection(ctx context.Context, serviceName string, registry Registry) (*grpc.ClientConn, error) {
	// Look up the IP:Port of all healthy instances of the requested service
	addrs, err := registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	log.Printf("Discovered %d instances of %s", len(addrs), serviceName)

	// Dial the target service by picking a random address from the discovery list.
	// grpc.WithStatsHandler injects the OpenTelemetry client handler for distributed tracing.
	return grpc.Dial(
		addrs[rand.Intn(len(addrs))],
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
}
