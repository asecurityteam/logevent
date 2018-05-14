package http

import (
	"io"
	"net/http"
	"os"
	"strings"

	"bitbucket.org/atlassian/logevent"
	"github.com/rs/xlog"
)

// levelFromString converts a string log level name into an xlog.Level type
// for use with xlog.
func levelFromString(level string) xlog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return xlog.LevelDebug
	case "INFO":
		return xlog.LevelInfo
	case "WARN":
		return xlog.LevelWarn
	case "ERROR":
		return xlog.LevelError
	case "FATAL":
		return xlog.LevelFatal
	default:
		return xlog.LevelDebug
	}
}

// Config is a data bag used to store the various state set by Options.
type Config struct {
	level         string
	humanReadable bool
	outputStream  io.Writer
}

// An Option configureds the middleware with some setting.
type Option func(*Config) *Config

// OptionLevel determines which logs are emitted based on level.
// The default values is INFO. Acceptable are ERROR, WARN, INFO, and DEBUG.
func OptionLevel(level string) func(*Config) *Config {
	return func(c *Config) *Config {
		c.level = level
		return c
	}
}

// OptionHumanReadable configures the logs for a more human readable format
// than the default JSON encoding. Note: Do no rely on the specific formatting
// of this option for automated processes. It is specifically designed for
// human readability and may change over time. Use the default JSON encoding
// for machine reading.
func OptionHumanReadable(c *Config) *Config {
	c.humanReadable = true
	return c
}

// OptionOutput customises the io.Writer used to ship all logs. The default
// value is os.Stdout.
func OptionOutput(output io.Writer) func(*Config) *Config {
	return func(c *Config) *Config {
		c.outputStream = output
		return c
	}
}

// Middleware wraps an http.Handler and injects a logevent.Logger in to the
// context.
type Middleware struct {
	wrapped http.Handler
}

// NewMiddleware generates an HTTP middleware with the given options set.
func NewMiddleware(options ...Option) (func(http.Handler) http.Handler, logevent.Logger) {
	var conf = &Config{
		level:        "INFO",
		outputStream: os.Stdout,
	}
	for _, option := range options {
		conf = option(conf)
	}
	var outputStream = xlog.NewConsoleOutputW(os.Stdout, xlog.NewLogfmtOutput(conf.outputStream))
	if !conf.humanReadable {
		outputStream = xlog.NewJSONOutput(conf.outputStream)
	}
	var c = xlog.Config{
		Level: levelFromString(conf.level),
		Output: xlog.OutputFunc(func(fields map[string]interface{}) error {
			return outputStream.Write(fields)
		}),
		DisablePooling: true,
	}
	var xlogMiddleware = xlog.NewHandler(c)
	return func(wrapped http.Handler) http.Handler {
		return xlogMiddleware(&Middleware{wrapped: wrapped})
	}, logevent.AdaptXlog(xlog.New(c))
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(logevent.NewContext(r.Context(), logevent.AdaptXlog(xlog.FromContext(r.Context()))))
	m.wrapped.ServeHTTP(w, r)
}

// FromRequest is a helper for extracting the logger from an *http.Request.
func FromRequest(r *http.Request) logevent.Logger {
	return logevent.FromContext(r.Context())
}
