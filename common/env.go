package common

import (
	"syscall"
)

// EnvString retrieves the value of the environment variable named by the key.
// It uses syscall.Getenv to access the underlying OS environment.
// If the variable is present in the environment, its value is returned.
// Otherwise, the provided fallback string is returned.
func EnvString(key, fallback string) string {
	// Attempt to get the environment variable directly from the OS
	if val, ok := syscall.Getenv(key); ok {
		return val
	}

	// Return the fallback value if the key wasn't found
	return fallback
}
