package logevent

import "context"

type ctxKey string

const (
	logeventKey = ctxKey("__logevent_ctx_key")
)

// NewContext installs a Logger.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, logeventKey, logger)
}

// FromContext fetches a Logger.
func FromContext(ctx context.Context) Logger {
	return ctx.Value(logeventKey).(Logger)
}
