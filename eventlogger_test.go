package logevent

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/fatih/structs"
	"github.com/golang/mock/gomock"
	"github.com/rs/xlog"
)

type eventNoMessage struct{}

type eventMessageNoAnnotations struct {
	Message string
}

type eventMessageAnnotationsNoDefault struct {
	Message string `logevent:"message"`
}

type eventMessageWrongType struct {
	Message int
}

type eventMessageBadAnnotation struct {
	Message string `logevent:"message,default"`
}

type eventMessage struct {
	One     string `logevent:"one"`
	Two     int    `logevent:"two"`
	Message string `logevent:"message,default=testvalue"`
}

type eventDefaultNumbers struct {
	Three   int     `logevent:"three,default=12"`
	Four    float64 `logevent:"four,default=.5"`
	Message string  `logevent:"message,default=testvalue"`
}

type EmbeddedStruct struct {
	Message string `logevent:"message,default=testvalue"`
	One     string `logevent:"one"`
}

type EventWithEmbeddedStructs struct {
	EmbeddedStruct
}

func TestLoggerWrapsContext(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var wrapped = newMockLogger(ctrl)
	var ctx = xlog.NewContext(context.Background(), wrapped)
	var r, _ = http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(ctx)

	_ = FromContext(ctx).(*logger)

}

func TestLoggerEventNoMessage(t *testing.T) {
	var s = structs.New(eventNoMessage{})
	var result = getMessage(s)
	if result != unknown {
		t.Fatalf("expected %s but got %s", unknown, result)
	}
}

func TestLoggerEventWrongMessageType(t *testing.T) {
	var s = structs.New(eventMessageWrongType{})
	var result = getMessage(s)
	if result != unknown {
		t.Fatalf("expected %s but got %s", unknown, result)
	}
}

func TestLoggerEventExplicitMessage(t *testing.T) {
	var s = structs.New(eventMessage{Message: "explicit"})
	var result = getMessage(s)
	if result != "explicit" {
		t.Fatalf("expected explicit but got %s", result)
	}
}

func TestLoggerEventEmptyMessageNoDefault(t *testing.T) {
	var s = structs.New(eventMessageAnnotationsNoDefault{})
	var result = getMessage(s)
	if result != unknown {
		t.Fatalf("expected %s but got %s", unknown, result)
	}
}

func TestLoggerEventEmptyMessageBadDefault(t *testing.T) {
	var s = structs.New(eventMessageBadAnnotation{})
	var result = getMessage(s)
	if result != unknown {
		t.Fatalf("expected %s but got %s", unknown, result)
	}
}

func TestLoggerEventEmptyMessageDefault(t *testing.T) {
	var s = structs.New(eventMessage{})
	var result = getMessage(s)
	if result != "testvalue" {
		t.Fatalf("expected testvalue but got %s", result)
	}
}

type defaultValueTestCase struct {
	TestValue   interface{}
	StringValue string
}

func TestLoggerEventDefaultValues(t *testing.T) {
	var s = structs.New(eventDefaultNumbers{})
	var intResult = getValue(s.Field("Three")).(int)
	if intResult != 12 {
		t.Fatalf("expected 12 but got %d", intResult)
	}
	var floatResult = getValue(s.Field("Four")).(float64)
	if floatResult != .5 {
		t.Fatalf("expected .5 but got %f", floatResult)
	}

	var cases = []defaultValueTestCase{
		defaultValueTestCase{TestValue: int(5), StringValue: "5"},
		defaultValueTestCase{TestValue: int8(5), StringValue: "5"},
		defaultValueTestCase{TestValue: int16(5), StringValue: "5"},
		defaultValueTestCase{TestValue: int32(5), StringValue: "5"},
		defaultValueTestCase{TestValue: int64(5), StringValue: "5"},
		defaultValueTestCase{TestValue: uint(5), StringValue: "5"},
		defaultValueTestCase{TestValue: uint8(5), StringValue: "5"},
		defaultValueTestCase{TestValue: uint16(5), StringValue: "5"},
		defaultValueTestCase{TestValue: uint32(5), StringValue: "5"},
		defaultValueTestCase{TestValue: uint64(5), StringValue: "5"},
		defaultValueTestCase{TestValue: float32(5.5), StringValue: "5.5"},
		defaultValueTestCase{TestValue: float64(5.5), StringValue: "5.5"},
		defaultValueTestCase{TestValue: string("5.5"), StringValue: "5.5"},
		defaultValueTestCase{TestValue: bool(true), StringValue: "true"},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("%s", reflect.TypeOf(testCase.TestValue)), func(tt *testing.T) {
			var s = structs.New(testCase)
			var result = getDefaultValue(s.Field("TestValue"), testCase.StringValue)
			if reflect.TypeOf(result) != reflect.TypeOf(testCase.TestValue) {
				tt.Errorf("failed to return correct type. instead got %s", reflect.TypeOf(result))
			}
			if result != testCase.TestValue {
				tt.Errorf("failed to extract string value of %s to %v. instead got %v", testCase.StringValue, testCase.TestValue, result)
			}
		})
	}
}

func TestLoggerTagsWithEventAttributesDebug(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = eventMessage{One: "one", Two: 2, Message: "testmessage"}
	var wrapped = newMockLogger(ctrl)
	var ctx = xlog.NewContext(context.Background(), wrapped)
	var logger = logger{logWithXlog, ctx}

	wrapped.EXPECT().OutputF(xlog.LevelDebug, 4, "testmessage", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		if _, ok = f["one"]; !ok {
			t.Fatal("missing attribute one")
		}
		if _, ok = f["two"]; !ok {
			t.Fatal("missing attribute two")
		}
	})
	logger.Debug(event)
}

func TestLoggerTagsWithEventAttributesInfo(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = eventMessage{One: "one", Two: 2, Message: "testmessage"}
	var wrapped = newMockLogger(ctrl)
	var ctx = xlog.NewContext(context.Background(), wrapped)
	var logger = logger{logWithXlog, ctx}

	wrapped.EXPECT().OutputF(xlog.LevelInfo, 4, "testmessage", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		if _, ok = f["one"]; !ok {
			t.Fatal("missing attribute one")
		}
		if _, ok = f["two"]; !ok {
			t.Fatal("missing attribute two")
		}
	})
	logger.Info(event)
}

func TestLoggerTagsWithEventAttributesWarn(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = eventMessage{One: "one", Two: 2, Message: "testmessage"}
	var wrapped = newMockLogger(ctrl)
	var ctx = xlog.NewContext(context.Background(), wrapped)
	var logger = logger{logWithXlog, ctx}

	wrapped.EXPECT().OutputF(xlog.LevelWarn, 4, "testmessage", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		if _, ok = f["one"]; !ok {
			t.Fatal("missing attribute one")
		}
		if _, ok = f["two"]; !ok {
			t.Fatal("missing attribute two")
		}
	})
	logger.Warn(event)
}

func TestLoggerTagsWithEventAttributesError(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = eventMessage{One: "one", Two: 2, Message: "testmessage"}
	var wrapped = newMockLogger(ctrl)
	var ctx = xlog.NewContext(context.Background(), wrapped)
	var logger = logger{logWithXlog, ctx}

	wrapped.EXPECT().OutputF(xlog.LevelError, 4, "testmessage", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		if _, ok = f["one"]; !ok {
			t.Fatal("missing attribute one")
		}
		if _, ok = f["two"]; !ok {
			t.Fatal("missing attribute two")
		}
	})
	logger.Error(event)
}

func TestLoggerTagsWithEmbeddedStructs(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var event = EventWithEmbeddedStructs{
		EmbeddedStruct: EmbeddedStruct{One: "one"},
	}
	var wrapped = newMockLogger(ctrl)
	var ctx = xlog.NewContext(context.Background(), wrapped)
	var logger = logger{logWithXlog, ctx}

	wrapped.EXPECT().OutputF(xlog.LevelError, 4, "testvalue", gomock.Any()).Do(func(l xlog.Level, c int, m string, f map[string]interface{}) {
		var ok bool
		if _, ok = f["one"]; !ok {
			t.Fatalf("missing attribute one, %v", f)
		}
	})
	logger.Error(event)
}
