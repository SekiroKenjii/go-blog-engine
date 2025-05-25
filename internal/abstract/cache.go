package abstract

import "context"

type ICacheService interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, any, int) error
	Delete(context.Context, string) error
	Exists(context.Context, string) (bool, error)
	Clear(context.Context) error
}
