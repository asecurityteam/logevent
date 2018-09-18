package logevent

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/structs"
	"github.com/rs/xlog"
)

type fallbackEvent struct {
	Message string `logevent:"message"`
}

type xlogLogger struct {
	xlogger xlog.Logger
	fields  *sync.Map
}

// Config records the requested settings for a logger for use with New().
type Config struct {
	// Level at which to log. Defaults to DEBUG.
	// Acceptable are ERROR, WARN, INFO, and DEBUG.
	Level string
	// HumanReadable toggles the JSON format off in favour of a colourised
	// log formatted for human readers.
	HumanReadable bool
	// Output defines to where logs are written. The default is os.Stdout.
	Output io.Writer
}

// New creates an instance of a Logger using the default backend.
func New(c Config) Logger {
	if c.Output == nil {
		c.Output = os.Stdout
	}
	var outputStream = xlog.NewConsoleOutputW(os.Stdout, xlog.NewLogfmtOutput(c.Output))
	if !c.HumanReadable {
		outputStream = xlog.NewJSONOutput(c.Output)
	}
	var xc = xlog.Config{
		Level: levelFromString(c.Level),
		Output: xlog.OutputFunc(func(fields map[string]interface{}) error {
			return outputStream.Write(fields)
		}),
		DisablePooling: true,
	}
	return &xlogLogger{xlogger: xlog.New(xc), fields: &sync.Map{}}
}

// Debug will emit the event with level DEBUG.
func (log *xlogLogger) Debug(event interface{}) {
	log.emit(xlog.LevelDebug, event)
}

// Info will emit the event with level INFO.
func (log *xlogLogger) Info(event interface{}) {
	log.emit(xlog.LevelInfo, event)
}

// Warn will emit the event with level WARN
func (log *xlogLogger) Warn(event interface{}) {
	log.emit(xlog.LevelWarn, event)
}

// Error will emit the event with level ERROR.
func (log *xlogLogger) Error(event interface{}) {
	log.emit(xlog.LevelError, event)
}

func (log *xlogLogger) emitString(level xlog.Level, event string) {
	log.emitStruct(level, fallbackEvent{Message: event})
}

func (log *xlogLogger) emitStruct(level xlog.Level, event interface{}) {
	var s = structs.New(event)
	var annotations = make(map[string]interface{})
	buildAnnotations(s, annotations)

	// apply logger level annotations, but don't override what was logged in a struct
	log.fields.Range(func(key interface{}, value interface{}) bool {
		addIfNotExists(annotations, key.(string), value)
		return true
	})

	var message = getMessage(s)
	delete(annotations, "message")
	log.xlogger.OutputF(level, 5, message, annotations)
}

func (log *xlogLogger) emit(level xlog.Level, event interface{}) {
	// Fallback for string values to unstructured logging. This exists to
	// help with migration paths from unstructure to structured by allowing
	// refactors to occur over time. It is **not** recommended to use this
	// feature if the logs can be made into structs.
	if msg, ok := event.(string); ok {
		log.emitString(level, msg)
		return
	}
	log.emitStruct(level, event)
}

// SetField applies a contextual annotation to all future events logged with
// this logger. This covers special cases where the annotations are either
// 1) not directly related to the event (such as logging context propagation
// from a remote call) or 2) part of an emerging set of common keys that would
// eventually be added automatically to structs via a request context.
func (log *xlogLogger) SetField(name string, value interface{}) {
	log.fields.Store(name, value)
}

// Copy the logger of use in some other context.
func (log *xlogLogger) Copy() Logger {
	var copy = &xlogLogger{
		xlogger: xlog.Copy(log.xlogger),
		fields:  &sync.Map{},
	}
	log.fields.Range(func(key interface{}, value interface{}) bool {
		copy.fields.Store(key, value)
		return true
	})
	return copy
}

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
