package logevent

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/fatih/structs"
	"github.com/rs/zerolog"
)

type fallbackEvent struct {
	Message string `logevent:"message"`
}

type logger struct {
	c      Config
	l      slog.Logger
	fields *sync.Map
}

// Config records the requested settings for a logger for use with New().
type Config struct {
	// Level at which to log. Defaults to DEBUG.
	// Acceptable are ERROR, WARN, INFO, and DEBUG.
	Level string
	// HumanReadable toggles the JSON format off in favor of a colorised
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
	ll := slog.New(slog.NewJSONHandler(c.Output,
		&slog.HandlerOptions{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}))
	zerolog.CallerSkipFrameCount = 5
	if c.HumanReadable {
		//FIXME add support for text log
		ll.Log(nil, slog.LevelWarn, "human readable log requested but not supported")
	}
	return &logger{
		c:      c,
		l:      *ll,
		fields: &sync.Map{},
	}
}

// Debug will emit the event with level DEBUG.
func (log *logger) Debug(event interface{}) {
	log.emit(slog.LevelDebug, event)
}

// Info will emit the event with level INFO.
func (log *logger) Info(event interface{}) {
	log.emit(slog.LevelInfo, event)
}

// Warn will emit the event with level WARN
func (log *logger) Warn(event interface{}) {
	log.emit(slog.LevelWarn, event)
}

// Error will emit the event with level ERROR.
func (log *logger) Error(event interface{}) {
	log.emit(slog.LevelError, event)
}

func (log *logger) emitString(level slog.Level, event string) {
	log.emitStruct(level, fallbackEvent{Message: event})
}

func (log *logger) emitStruct(level slog.Level, event interface{}) {
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
	if message == unknown {
		// struct is lacking a Message field, or Message field is "".
		// As a last resort, see if the event is error type
		if _, ok := event.(error); ok {
			message = event.(error).Error()
		}
	}
	//FIXME annotations
	_, file, line, _ := runtime.Caller(4)
	log.l.Log(context.Background(), level, message, "file", file, "line", line)
}

func (log *logger) emit(level slog.Level, event interface{}) {
	// Fallback for string values to unstructured logging. This exists to
	// help with migration paths from unstructured to structured by allowing
	// refactors to occur over time. It is **not** recommended to use this
	// feature if the logs can be made into structs.
	if event == nil {
		event = "(nil)"
	}
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
	var cp = New(log.c).(*logger)
	log.fields.Range(func(key interface{}, value interface{}) bool {
		cp.fields.Store(key, value)
		return true
	})
	return cp
}

// levelFromString converts a string log level name into an slog.Level type
// for use with slog.
func levelFromString(level string) slog.Level {
	var l slog.Level
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		return slog.LevelDebug
	}
	return l
}
