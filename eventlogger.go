package logevent

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/rs/xlog"
)

const (
	tagKey       = "logevent"
	unknown      = "unknown"
	defaultValue = "default="
)

// LogLevel represents a levelled logging string that indicates the priority
// of the event.
type LogLevel string

var (
	// LogLevelDebug is used to indicate DEBUG priority events.
	LogLevelDebug LogLevel = "DEBUG"
	// LogLevelInfo is used to indicate INFO priority events.
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarn is used to indicate WARN priority events.
	LogLevelWarn LogLevel = "WARN"
	// LogLevelError is used to indicate ERROR priority events.
	LogLevelError LogLevel = "ERROR"
)

// LogFunc is the primary abstraction point for new implementations. This will
// be called for each event once it has been decomposed into components.
type LogFunc func(ctx context.Context, level LogLevel, message string, annotations map[string]interface{})

// Logger represents a level based log emitter that operates on event structs.
type Logger interface {
	// Debug will emit the event with level DEBUG.
	Debug(interface{})
	// Info will emit the event with level INFO.
	Info(interface{})
	// Warn will emit the event with level WARN
	Warn(interface{})
	// Error will emit the event with level ERROR.
	Error(interface{})
}

func logWithXlog(ctx context.Context, level LogLevel, message string, annotations map[string]interface{}) {
	var xlogLevel xlog.Level
	switch level {
	case LogLevelDebug:
		xlogLevel = xlog.LevelDebug
	case LogLevelInfo:
		xlogLevel = xlog.LevelInfo
	case LogLevelWarn:
		xlogLevel = xlog.LevelWarn
	case LogLevelError:
		xlogLevel = xlog.LevelError
	default:
		xlogLevel = xlog.LevelInfo
	}
	xlog.FromContext(ctx).OutputF(xlogLevel, 4, message, annotations)
}

type loggerProvider struct {
	f LogFunc
}

func (p *loggerProvider) fromContext(ctx context.Context) Logger {
	return New(ctx, p.f)
}

var defaultProvider = &loggerProvider{logWithXlog}

// FromContext will fetch an xlog.Logger from the context, wrap it in the
// Logger interface, and return the wrapped logger.
var FromContext = defaultProvider.fromContext

// NewFromContextFunc provides a FromContext equivalent for any LogFunc.
func NewFromContextFunc(f LogFunc) func(context.Context) Logger {
	return (&loggerProvider{f}).fromContext
}

// New wraps a given LogFunc in an event based Logger interface.
func New(ctx context.Context, f LogFunc) Logger {
	return &logger{wrapped: f, ctx: ctx}
}

func getDefaultValue(f *structs.Field, value string) interface{} {
	switch reflect.TypeOf(f.Value()).Kind() {
	case reflect.String:
		return value
	case reflect.Bool:
		var final, _ = strconv.ParseBool(value)
		return final
	case reflect.Int:
		var final, _ = strconv.ParseInt(value, 10, strconv.IntSize)
		return int(final)
	case reflect.Int8:
		var final, _ = strconv.ParseInt(value, 10, 8)
		return int8(final)
	case reflect.Int16:
		var final, _ = strconv.ParseInt(value, 10, 16)
		return int16(final)
	case reflect.Int32:
		var final, _ = strconv.ParseInt(value, 10, 32)
		return int32(final)
	case reflect.Int64:
		var final, _ = strconv.ParseInt(value, 10, 64)
		return int64(final)
	case reflect.Float32:
		var final, _ = strconv.ParseFloat(value, 32)
		return float32(final)
	case reflect.Float64:
		var final, _ = strconv.ParseFloat(value, 64)
		return float64(final)
	default:
		return f.Value()
	}
}

func getName(f *structs.Field) string {
	var tags = strings.Split(f.Tag(tagKey), ",")
	if len(tags) < 1 {
		return f.Name()
	}
	return tags[0]
}

func getValue(f *structs.Field) interface{} {
	if !f.IsZero() {
		return f.Value()
	}
	var tags = strings.Split(f.Tag(tagKey), ",")
	for _, tag := range tags {
		if strings.Contains(tag, defaultValue) {
			var parts = strings.Split(tag, "=")
			if len(parts) == 2 {
				return getDefaultValue(f, parts[1])
			}
		}
	}
	return f.Value()
}

func getMessage(s *structs.Struct) string {
	var message string
	var msgField *structs.Field
	var ok bool
	msgField, ok = s.FieldOk("Message")
	if !ok {
		return unknown
	}
	message, ok = getValue(msgField).(string)
	if ok && len(message) > 0 {
		return message
	}
	return unknown
}

type logger struct {
	wrapped LogFunc
	ctx     context.Context
}

func (l *logger) emit(level LogLevel, event interface{}) {
	var s = structs.New(event)
	s.TagName = tagKey
	var strucs = []*structs.Struct{s}
	var annotations = make(map[string]interface{})
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
	l.wrapped(l.ctx, level, message, annotations)
}

func (l *logger) Debug(event interface{}) {
	l.emit(LogLevelDebug, event)
}
func (l *logger) Info(event interface{}) {
	l.emit(LogLevelInfo, event)
}
func (l *logger) Warn(event interface{}) {
	l.emit(LogLevelWarn, event)
}
func (l *logger) Error(event interface{}) {
	l.emit(LogLevelError, event)
}
