package logevent

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/xlog"
	"github.com/stretchr/testify/require"
)

type tagTestCase struct {
	Level xlog.Level
	Func  func(eventMessage, Logger)
}

func TestLoggerTagsWithEventAttributesLevels(t *testing.T) {
	var cases = []tagTestCase{
		tagTestCase{Level: xlog.LevelDebug, Func: func(ev eventMessage, logger Logger) {
			logger.Debug(ev)
		}},
		tagTestCase{Level: xlog.LevelInfo, Func: func(ev eventMessage, logger Logger) {
			logger.Info(ev)
		}},
		tagTestCase{Level: xlog.LevelWarn, Func: func(ev eventMessage, logger Logger) {
			logger.Warn(ev)
		}},
		tagTestCase{Level: xlog.LevelError, Func: func(ev eventMessage, logger Logger) {
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
	Level xlog.Level
	Func  func(string, Logger)
}

func TestLoggerTagsStringWithAttributesLevels(t *testing.T) {
	var cases = []stringTagTestCase{
		stringTagTestCase{Level: xlog.LevelDebug, Func: func(ev string, logger Logger) {
			logger.Debug(ev)
		}},
		stringTagTestCase{Level: xlog.LevelInfo, Func: func(ev string, logger Logger) {
			logger.Info(ev)
		}},
		stringTagTestCase{Level: xlog.LevelWarn, Func: func(ev string, logger Logger) {
			logger.Warn(ev)
		}},
		stringTagTestCase{Level: xlog.LevelError, Func: func(ev string, logger Logger) {
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
	var event = EventWithEmbeddedStructs{}
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
}

func TestLoggerTagsWithNestedStructs(t *testing.T) {
	var nestedEvent = EventWithNestedStructs{
		Nested: EmbeddedStruct{One: "one"},
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
}

func TestLoggerTagsWithNestedEmbeddedStructs(t *testing.T) {
	var nestedEvent = EventWithNestedEmbeddedStructs{
		Nested: EventWithEmbeddedStructs{},
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

}
