package logevent

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/structs"
	"github.com/rs/zerolog"
)

type fallbackEvent struct {
	Message string `logevent:"message"`
}

type logger struct {
	c      Config
	l      zerolog.Logger
	fields *sync.Map
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
	zerolog.CallerFieldName = "file"
	zerolog.CallerSkipFrameCount = 5
	var l = zerolog.New(c.Output).With().Caller().Timestamp().Logger().Level(levelFromString(c.Level))
	if c.HumanReadable {
		l = l.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	return &logger{
		c:      c,
		l:      l,
		fields: &sync.Map{},
	}
}

// Debug will emit the event with level DEBUG.
func (log *logger) Debug(event interface{}) {
	log.emit(zerolog.DebugLevel, event)
}

// Info will emit the event with level INFO.
func (log *logger) Info(event interface{}) {
	log.emit(zerolog.InfoLevel, event)
}

// Warn will emit the event with level WARN
func (log *logger) Warn(event interface{}) {
	log.emit(zerolog.WarnLevel, event)
}

// Error will emit the event with level ERROR.
func (log *logger) Error(event interface{}) {
	log.emit(zerolog.ErrorLevel, event)
}

func (log *logger) emitString(level zerolog.Level, event string) {
	log.emitStruct(level, fallbackEvent{Message: event})
}

func (log *logger) emitStruct(level zerolog.Level, event interface{}) {
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
	log.l.WithLevel(level).Fields(annotations).Msg(message)
}

func (log *logger) emit(level zerolog.Level, event interface{}) {
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
func (log *logger) SetField(name string, value interface{}) {
	log.fields.Store(name, value)
}

// Copy the logger of use in some other context.
func (log *logger) Copy() Logger {
	var copy = New(log.c).(*logger)
	log.fields.Range(func(key interface{}, value interface{}) bool {
		copy.fields.Store(key, value)
		return true
	})
	return copy
}

// levelFromString converts a string log level name into an xlog.Level type
// for use with xlog.
func levelFromString(level string) zerolog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	default:
		return zerolog.DebugLevel
	}
}
