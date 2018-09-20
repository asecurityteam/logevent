package logevent

import (
	"fmt"
	"io/ioutil"
	"testing"
)

type Foo struct {
	Value  int         `logevent:"value"`
	Nested interface{} `logevent:"nested"`
}

const maxDepth = 24
const delta = 2

func BenchmarkLog(b *testing.B) {
	for depth := 0; depth <= maxDepth; depth += delta {
		var name = fmt.Sprintf("depth=%d", depth)
		b.Run(name, func(b *testing.B) {
			var event = buildEvent(depth)
			var logger = New(Config{Output: ioutil.Discard})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Error(*event)
			}
		})
	}
}

func buildEvent(depth int) *Foo {
	var event = &Foo{Value: 0}
	var prev = event
	for currentDepth := 1; currentDepth < depth; currentDepth++ {
		var newValue = &Foo{
			Value: currentDepth,
		}
		prev.Nested = newValue
		prev = prev.Nested.(*Foo)
	}
	return event
}
