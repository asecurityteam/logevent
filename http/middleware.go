package http

import (
	"net/http"

	"bitbucket.org/atlassian/logevent"
)

// Middleware wraps an http.Handler and injects a logevent.Logger in to the
// context.
type Middleware struct {
	logger  logevent.Logger
	wrapped http.Handler
}

// NewMiddleware generates an HTTP middleware with the given options set.
func NewMiddleware(logger logevent.Logger) func(http.Handler) http.Handler {
	return func(wrapped http.Handler) http.Handler {
		return &Middleware{wrapped: wrapped, logger: logger}
	}
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(logevent.NewContext(r.Context(), m.logger.Copy()))
	m.wrapped.ServeHTTP(w, r)
}

// FromRequest is a helper for extracting the logger from an *http.Request.
func FromRequest(r *http.Request) logevent.Logger {
	return logevent.FromContext(r.Context())
}
