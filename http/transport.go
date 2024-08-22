package http

import (
	"net/http"

	"github.com/asecurityteam/logevent/v2"
)

// Transport injects a `logevent.Logger` into the request context
// during its `http.RoundTrip`.
type Transport struct {
	logger  logevent.Logger
	wrapped http.RoundTripper
}

// RoundTrip injects a `logevent.Logger` into the current request context.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.WithContext(logevent.NewContext(r.Context(), t.logger.Copy()))
	return t.wrapped.RoundTrip(r)
}

// NewTransport wraps a `transport.Decorator` in a new one that injects a
// `logevent.Logger` into the context.
func NewTransport(logger logevent.Logger) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &Transport{logger: logger, wrapped: next}
	}
}
