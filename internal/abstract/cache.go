package abstract

import (
	"context"
	"time"
)

type ICacheService interface {
	// Get retrieves a value from the cache by its key.
	Get(context.Context, string) (string, error)

	// GetBit retrieves a bit value from a string at a specific offset.
	GetBit(context.Context, string, int) (int64, error)

	// Set stores a value in the cache with a specified expiration time.
	Set(context.Context, string, any, int) error

	// SetBit sets a bit value in a string at a specific offset.
	SetBit(context.Context, string, int, int) (int64, error)

	// Delete removes a value from the cache by its key.
	Delete(context.Context, string) error

	// Exists checks if a key exists in the cache.
	Exists(context.Context, string) (bool, error)

	// Clear removes all values from the cache.
	Clear(context.Context) error

	// GetWithDefault retrieves a value from the cache or returns a default value if not found.
	GetWithDefault(context.Context, string, string) (string, error)

	// SetWithDefaultExpiration sets a value in the cache with a default expiration time.
	SetWithDefaultExpiration(context.Context, string, any) error

	// SetNX sets a value in the cache only if the key does not already exist.
	SetNX(context.Context, string, any, int) (bool, error)

	// GetTTL retrieves the time-to-live for a key in the cache.
	GetTTL(context.Context, string) (time.Duration, error)

	// Increment increments the value of a key in the cache by 1.
	Increment(context.Context, string) (int64, error)

	// IncrementBy increments the value of a key in the cache by a specified amount.
	IncrementBy(context.Context, string, int64) (int64, error)

	// Decrement decrements the value of a key in the cache by 1.
	Keys(context.Context, string) ([]string, error)
}
