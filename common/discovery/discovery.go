package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Registry defines the standard interface for a service discovery mechanism (e.g., Consul).
// It allows microservices to register themselves on startup and discover other services 
// dynamically without hardcoding IP addresses.
type Registry interface {
	// Register adds a new service instance to the registry.
	Register(ctx context.Context, instanceID, serviceName, hostPort string) error
	
	// Deregister removes a service instance from the registry (typically on shutdown).
	Deregister(ctx context.Context, instanceID, serviceName string) error
	
	// Discover looks up the active instances for a given service name and returns their addresses.
	Discover(ctx context.Context, serviceName string) ([]string, error)
	
	// HealthCheck allows the registry to ping the service to ensure it's still alive.
	HealthCheck(instanceID, serviceName string) error
}

// GenerateInstanceID creates a unique identifier for a specific running instance of a microservice.
// This is used to differentiate multiple replicas of the same service (e.g., 3 instances of 'orders').
func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
