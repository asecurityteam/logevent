package logevent

// Logger is a logging system abstraction that supports leveled, strictly
// structured log emissions.
type Logger interface {
	// Debug will emit the event with level DEBUG.
	Debug(event interface{})
	// Info will emit the event with level INFO.
	Info(event interface{})
	// Warn will emit the event with level WARN
	Warn(event interface{})
	// Error will emit the event with level ERROR.
	Error(event interface{})
	// SetField applies a contextual annotation to all
	// future events logged with this logger.
	SetField(name string, value interface{})
	// Copy the logger of use in some other context.
	Copy() Logger
}
