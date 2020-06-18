package logevent

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var timeField = time.Now()

type EmbeddedStruct struct {
	Message string    `logevent:"message,default=testvalue"`
	One     string    `logevent:"one,default=foo"`
	Two     time.Time `logevent:"two"`
}

type EventWithEmbeddedStructs struct {
	EmbeddedStruct
	One string `logevent:"one,default=fizz"`
}

type EventWithNestedStructs struct {
	Message string         `logevent:"message,default=testvalue"`
	Nested  EmbeddedStruct `logevent:"nested"`
}

type EventWithDoubleNestedStructs struct {
	Message string                 `logevent:"message,default=testvalue"`
	Nested  EventWithNestedStructs `logevent:"nested"`
}

type EventWithNestedEmbeddedStructs struct {
	Message string                   `logevent:"message,default=testvalue"`
	Nested  EventWithEmbeddedStructs `logevent:"nested"`
}

type EventWithUnexportedField struct {
	Message    string `logevent:"message,default=testvalue"`
	unexported string `logevent:"unexported,default=fizz"`
}

type EventWithNoField struct {
}

type tagTestCase struct {
	Level zerolog.Level
	Func  func(eventMessage, Logger)
}

func TestLoggerTagsWithEventAttributesLevels(t *testing.T) {
	var cases = []tagTestCase{
		{Level: zerolog.DebugLevel, Func: func(ev eventMessage, logger Logger) {
			logger.Debug(ev)
		}},
		{Level: zerolog.InfoLevel, Func: func(ev eventMessage, logger Logger) {
			logger.Info(ev)
		}},
		{Level: zerolog.WarnLevel, Func: func(ev eventMessage, logger Logger) {
			logger.Warn(ev)
		}},
		{Level: zerolog.ErrorLevel, Func: func(ev eventMessage, logger Logger) {
			logger.Error(ev)
		}},
	}
	for _, currentCase := range cases {
		t.Run(string(currentCase.Level), func(tb *testing.T) {
			var event = eventMessage{One: "one", Two: 2, Message: "testmessage"}
			var buff = &bytes.Buffer{}
			var c = Config{Output: buff}
			var logger = New(c)
			logger.SetField("out-of-event", "true")
			currentCase.Func(event, logger)

			// trim the extra empty line, and split all lines
			var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
			var line = make(map[string]interface{})
			_ = json.Unmarshal([]byte(lines[0]), &line)
			var _, okFile = line["file"]
			var _, okTime = line["time"]
			require.True(t, okFile, "log line missing file attribute")
			require.True(t, okTime, "log line missing time attribute")
			require.Equal(t, currentCase.Level, levelFromString(line["level"].(string)))
			require.Equal(t, "testmessage", line["message"])
			require.Equal(t, "one", line["one"])
			require.Equal(t, "true", line["out-of-event"])
			require.Equal(t, 2.0, line["two"])
		})
	}
}

type stringTagTestCase struct {
	Level zerolog.Level
	Func  func(string, Logger)
}

func TestLoggerTagsStringWithAttributesLevels(t *testing.T) {
	var cases = []stringTagTestCase{
		{Level: zerolog.DebugLevel, Func: func(ev string, logger Logger) {
			logger.Debug(ev)
		}},
		{Level: zerolog.InfoLevel, Func: func(ev string, logger Logger) {
			logger.Info(ev)
		}},
		{Level: zerolog.WarnLevel, Func: func(ev string, logger Logger) {
			logger.Warn(ev)
		}},
		{Level: zerolog.ErrorLevel, Func: func(ev string, logger Logger) {
			logger.Error(ev)
		}},
	}
	for _, currentCase := range cases {
		t.Run(string(currentCase.Level), func(tb *testing.T) {
			var event = "testmessage"
			var buff = &bytes.Buffer{}
			var c = Config{Output: buff}
			var logger = New(c)
			logger.SetField("out-of-event", "true")
			currentCase.Func(event, logger)

			// trim the extra empty line, and split all lines
			var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
			var line = make(map[string]interface{})
			_ = json.Unmarshal([]byte(lines[0]), &line)
			var _, okFile = line["file"]
			var _, okTime = line["time"]
			require.True(t, okFile, "log line missing file attribute")
			require.True(t, okTime, "log line missing time attribute")
			require.Equal(t, currentCase.Level, levelFromString(line["level"].(string)))
			require.Equal(t, "testmessage", line["message"])
			require.Equal(t, "true", line["out-of-event"])
		})
	}
}

func TestLoggerTagsWithEmbeddedStructs(t *testing.T) {
	var embeddedStruct = EmbeddedStruct{Two: timeField}
	var event = EventWithEmbeddedStructs{EmbeddedStruct: embeddedStruct}
	var buff = &bytes.Buffer{}
	var c = Config{Output: buff}
	var logger = New(c)
	logger.SetField("one", "override me!")
	logger.Error(event)

	// trim the extra empty line, and split all lines
	var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
	var line = make(map[string]interface{})
	_ = json.Unmarshal([]byte(lines[0]), &line)
	var _, okFile = line["file"]
	var _, okTime = line["time"]
	require.True(t, okFile, "log line missing file attribute")
	require.True(t, okTime, "log line missing time attribute")
	require.Equal(t, "error", line["level"])
	require.Equal(t, "testvalue", line["message"])
	require.Equal(t, "fizz", line["one"])
	require.Equal(t, timeField.Format(time.RFC3339Nano), line["two"])
}

func TestLoggerTagsWithNestedStructs(t *testing.T) {
	var nestedEvent = EventWithNestedStructs{
		Nested: EmbeddedStruct{
			One: "one",
			Two: timeField,
		},
	}
	var doubleNestedEvent = EventWithDoubleNestedStructs{
		Nested: nestedEvent,
	}
	var buff = &bytes.Buffer{}
	var c = Config{Output: buff}
	var logger = New(c)

	logger.Error(doubleNestedEvent)

	var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
	var line = make(map[string]interface{})
	_ = json.Unmarshal([]byte(lines[0]), &line)
	var _, okFile = line["file"]
	var _, okTime = line["time"]
	require.True(t, okFile, "log line missing file attribute")
	require.True(t, okTime, "log line missing time attribute")
	require.Equal(t, "error", line["level"])
	require.Equal(t, "testvalue", line["message"])

	var nested, okNested = line["nested"]
	require.True(t, okNested, "log line missing nested attribute")
	var nestedStruct = nested.(map[string]interface{})
	require.Equal(t, "testvalue", nestedStruct["message"])

	var doubleNested, okDoubleNested = nestedStruct["nested"]
	require.True(t, okDoubleNested, "log line missing nested attribute")
	var doubleNestedStruct = doubleNested.(map[string]interface{})
	require.Equal(t, "testvalue", doubleNestedStruct["message"])
	require.Equal(t, "one", doubleNestedStruct["one"])
	require.Equal(t, timeField.Format(time.RFC3339Nano), doubleNestedStruct["two"])
}

func TestLoggerTagsWithNestedEmbeddedStructs(t *testing.T) {
	var embeddedStruct = EventWithEmbeddedStructs{EmbeddedStruct: EmbeddedStruct{Two: timeField}}
	var nestedEvent = EventWithNestedEmbeddedStructs{
		Nested: embeddedStruct,
	}

	var buff = &bytes.Buffer{}
	var c = Config{Output: buff}
	var logger = New(c)

	logger.Error(nestedEvent)

	var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
	var line = make(map[string]interface{})
	_ = json.Unmarshal([]byte(lines[0]), &line)
	var _, okFile = line["file"]
	var _, okTime = line["time"]
	require.True(t, okFile, "log line missing file attribute")
	require.True(t, okTime, "log line missing time attribute")
	require.Equal(t, "error", line["level"])
	require.Equal(t, "testvalue", line["message"])

	var nested, okNested = line["nested"]
	require.True(t, okNested, "log line missing nested attribute")
	var nestedStruct = nested.(map[string]interface{})
	require.Equal(t, "testvalue", nestedStruct["message"])
	require.Equal(t, "fizz", nestedStruct["one"])
	require.Equal(t, timeField.Format(time.RFC3339Nano), nestedStruct["two"])
}

func TestLoggerTagsWithUnexportedField(t *testing.T) {
	var eventWithUnexported = EventWithUnexportedField{
		Message:    "testmessage",
		unexported: "unexported",
	}

	var buff = &bytes.Buffer{}
	var c = Config{Output: buff}
	var logger = New(c)

	logger.Error(eventWithUnexported)

	var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
	var line = make(map[string]interface{})
	_ = json.Unmarshal([]byte(lines[0]), &line)
	var _, okFile = line["file"]
	var _, okTime = line["time"]
	require.True(t, okFile, "log line missing file attribute")
	require.True(t, okTime, "log line missing time attribute")
	require.Equal(t, "error", line["level"])
	require.Equal(t, "testmessage", line["message"])

	var _, okExported = line["message"]
	require.True(t, okExported, "log line missing nested attribute")

	// unexported fields should be silently omitted from the log
	var _, okUnexported = line["unexported"]
	require.False(t, okUnexported, "log line has nested attribute")

}

func TestLoggerTagsWithNoFields(t *testing.T) {
	var eventWithNothing = EventWithNoField{}

	var buff = &bytes.Buffer{}
	var c = Config{Output: buff}
	var logger = New(c)

	logger.Error(eventWithNothing)

	var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
	var line = make(map[string]interface{})
	_ = json.Unmarshal([]byte(lines[0]), &line)
	var _, okFile = line["file"]
	var _, okTime = line["time"]
	require.True(t, okFile, "log line missing file attribute")
	require.True(t, okTime, "log line missing time attribute")
	require.Equal(t, "error", line["level"])
	require.Equal(t, "unknown", line["message"])

	var _, okExported = line["message"]
	require.True(t, okExported, "log line missing nested attribute")

}

func TestLoggerAccidentalNil(t *testing.T) {
	// sometimes people accidentally code `logger.Info(thing)` there thing == nil.  It happens.  Don't panic.

	var buff = &bytes.Buffer{}
	var c = Config{Output: buff}
	var logger = New(c)

	var err error
	logger.Error(err)

	var lines = strings.Split(strings.Trim(buff.String(), "\n"), "\n")
	var line = make(map[string]interface{})
	_ = json.Unmarshal([]byte(lines[0]), &line)
	var _, okFile = line["file"]
	var _, okTime = line["time"]
	require.True(t, okFile, "log line missing file attribute")
	require.True(t, okTime, "log line missing time attribute")
	require.Equal(t, "error", line["level"])
	require.Equal(t, "(nil)", line["message"])

	var _, okExported = line["message"]
	require.True(t, okExported, "log line missing nested attribute")
}
