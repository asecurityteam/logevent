package logevent

import (
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

// AdaptXlog creates a logevent instance from an xlog instance. This exists to
// support working with the default logging backend which is github.com/rs/xlog.
// However, one of the values of using this framework is that it 1) abstracts
// the underlying logging system used and 2) allows for custom implementations
// of the logevent.Logger. Only use this adapter if absolutely necessary.
func AdaptXlog(logger xlog.Logger) Logger {
	return &xlogLogger{xlogger: logger, fields: &sync.Map{}}
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
	s.TagName = tagKey
	var annotations = make(map[string]interface{})
	log.fields.Range(func(key interface{}, value interface{}) bool {
		annotations[key.(string)] = value
		return true
	})
	var strucs = []*structs.Struct{s}
	for len(strucs) > 0 {
		for _, field := range strucs[0].Fields() {
			if structs.IsStruct(field.Value()) && field.IsEmbedded() {
				strucs = append(strucs, structs.New(field.Value()))
				continue
			}
			annotations[getName(field)] = getValue(field)
		}
		strucs = strucs[1:]
	}

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
