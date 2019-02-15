package logevent

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/fatih/structs"
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
	One     string `logevent:"one,default=foo"`
	Two     int    `logevent:"two"`
	Message string `logevent:"message,default=testvalue"`
}

type eventDefaultNumbers struct {
	Three   int     `logevent:"three,default=12"`
	Four    float64 `logevent:"four,default=.5"`
	Message string  `logevent:"message,default=testvalue"`
}

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
