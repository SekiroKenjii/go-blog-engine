package abstract

import (
	"context"
	"time"
)

type ICacheService interface {
	Get(context.Context, string) (string, error)
	GetBit(context.Context, string, int) (int64, error)
	Set(context.Context, string, any, int) error
	SetBit(context.Context, string, int, int) (int64, error)
	Delete(context.Context, string) error
	Exists(context.Context, string) (bool, error)
	Clear(context.Context) error
	GetWithDefault(context.Context, string, string) (string, error)
	SetWithDefaultExpiration(context.Context, string, any) error
	SetNX(context.Context, string, any, int) (bool, error)
	GetTTL(context.Context, string) (time.Duration, error)
	Increment(context.Context, string) (int64, error)
	IncrementBy(context.Context, string, int64) (int64, error)
	Keys(context.Context, string) ([]string, error)
}
