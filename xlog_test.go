package logevent

import (
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/xlog"
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
			var ctrl = gomock.NewController(tb)
			defer ctrl.Finish()
			var event = eventMessage{One: "one", Two: 2, Message: "testmessage"}
			var wrapped = newMockLogger(ctrl)
			var logger = &xlogLogger{wrapped, &sync.Map{}}
			logger.SetField("out-of-event", "true")
			wrapped.EXPECT().OutputF(currentCase.Level, 5, "testmessage", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
				var ok bool
				if _, ok = f["one"]; !ok {
					t.Fatal("missing attribute one")
				}
				if _, ok = f["two"]; !ok {
					t.Fatal("missing attribute two")
				}
				if _, ok = f["out-of-event"]; !ok {
					t.Fatal("missing attribute out-of-event")
				}
			})
			currentCase.Func(event, logger)
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
			var ctrl = gomock.NewController(tb)
			defer ctrl.Finish()
			var event = "testmessage"
			var wrapped = newMockLogger(ctrl)
			var logger = &xlogLogger{wrapped, &sync.Map{}}
			logger.SetField("out-of-event", "true")
			wrapped.EXPECT().OutputF(currentCase.Level, 5, "testmessage", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
				var ok bool
				if _, ok = f["out-of-event"]; !ok {
					t.Fatal("missing attribute out-of-event")
				}
			})
			currentCase.Func(event, logger)
		})
	}
}

func TestLoggerTagsWithEmbeddedStructs(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = EventWithEmbeddedStructs{}
	var wrapped = newMockLogger(ctrl)
	var logger = &xlogLogger{wrapped, &sync.Map{}}
	logger.SetField("one", "override me!")

	wrapped.EXPECT().OutputF(xlog.LevelError, 5, "testvalue", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		var val interface{}
		if val, ok = f["one"]; !ok {
			t.Fatalf("missing attribute one, %v", f)
		} else if val != "fizz" {
			t.Fatalf("Expected default value of embedded struct field to be overridden to fizz, but was %v", val)
		}
	})
	logger.Error(event)
}

func TestLoggerTagsWithNestedStructs(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = EventWithNestedStructs{
		Nested: EmbeddedStruct{One: "one"},
	}
	var wrapped = newMockLogger(ctrl)
	var logger = &xlogLogger{wrapped, &sync.Map{}}

	wrapped.EXPECT().OutputF(xlog.LevelError, 5, "testvalue", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		if _, ok = f["nested"]; !ok {
			t.Fatalf("missing attribute nested, %v", f)
		}
		if _, ok = f["nested"].(EmbeddedStruct); !ok {
			t.Fatalf("nested attribute type was not correct, %v", f)
		}
		if f["nested"].(EmbeddedStruct).One != "one" {
			t.Fatalf("nested attribute value was not correct, %v", f)
		}
	})
	logger.Error(event)
}
