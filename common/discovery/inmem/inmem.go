package inmem

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Registry provides an in-memory implementation of the discovery.Registry interface.
// This is primarily useful for testing or local development when Consul isn't available.
type Registry struct {
	sync.RWMutex
	// addrs maps a service name to a map of instance IDs to their respective details
	addrs map[string]map[string]*serviceInstance
}

// serviceInstance holds the networking information and liveness state of a single service replica.
type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

// NewRegistry initializes an empty thread-safe in-memory registry.
func NewRegistry() *Registry {
	return &Registry{addrs: map[string]map[string]*serviceInstance{}}
}

// Register saves the service instance details into memory under its service name and instance ID.
func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	r.Lock()
	defer r.Unlock()

	// Initialize the map for the service name if it doesn't exist
	if _, ok := r.addrs[serviceName]; !ok {
		r.addrs[serviceName] = map[string]*serviceInstance{}
	}

	// Record the instance along with the current time as its last active timestamp
	r.addrs[serviceName][instanceID] = &serviceInstance{hostPort: hostPort, lastActive: time.Now()}

	return nil
}

// Deregister forcefully removes an instance from the registry.
func (r *Registry) Deregister(ctx context.Context, instanceID, serviceName string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.addrs[serviceName]; !ok {
		return nil
	}

	delete(r.addrs[serviceName], instanceID)

	return nil
}

// HealthCheck updates the lastActive timestamp for the instance, simulating a TTL refresh.
func (r *Registry) HealthCheck(instanceID, serviceName string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.addrs[serviceName]; !ok {
		return errors.New("service is not registered yet")
	}

	if _, ok := r.addrs[serviceName][instanceID]; !ok {
		return errors.New("service instance is not registered yet")
	}

	r.addrs[serviceName][instanceID].lastActive = time.Now()

	return nil
}

// Discover retrieves all registered addresses for a service, ignoring their health status.
// Contrast this with ServiceAddresses which filters out stale instances.
func (r *Registry) Discover(ctx context.Context, serviceName string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()

	if len(r.addrs[serviceName]) == 0 {
		return nil, errors.New("no service address found")
	}

	var res []string
	for _, i := range r.addrs[serviceName] {
		res = append(res, i.hostPort)
	}

	return res, nil
}

// ServiceAddresses returns a list of healthy instances for the given service.
// It filters out any instances that haven't called HealthCheck in the last 5 seconds.
func (r *Registry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()

	if len(r.addrs[serviceName]) == 0 {
		return nil, errors.New("no service address found")
	}

	var res []string
	for _, i := range r.addrs[serviceName] {
		// If the instance hasn't pinged back in 5 seconds, consider it dead
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			continue
		}
		res = append(res, i.hostPort)
	}

	return res, nil
}
