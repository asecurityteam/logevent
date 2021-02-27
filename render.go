package logevent

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/structs"
)

const (
	tagKey       = "logevent"
	unknown      = "unknown"
	defaultValue = "default="
)

var mutex = &sync.RWMutex{}

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
		return final
	case reflect.Float32:
		var final, _ = strconv.ParseFloat(value, 32)
		return float32(final)
	case reflect.Float64:
		var final, _ = strconv.ParseFloat(value, 64)
		return final
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

// getMessage will render the value of the unknown const
// if there is no Message field in the struct
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

func buildAnnotations(s *structs.Struct, annotations map[string]interface{}) {

	s.TagName = tagKey
	var strucs = []*structs.Struct{s}
	for len(strucs) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(strucs[0].Fields()))
		for _, field := range strucs[0].Fields() {
			// we spin up go routines to best-effort add annotations, but
			// catch panics if any attempt to get an unexported field occurs
			go func(field *structs.Field, annotations map[string]interface{}) {
				// if any panic occurs, we don't want to break the runtime caller, so recover
				defer func() {
					if err := recover(); err != nil {
						fmt.Println(err)
					}
					wg.Done()
				}()
				if structs.IsStruct(field.Value()) {
					var fieldStruct = structs.New(field.Value())
					if field.IsEmbedded() {
						strucs = append(strucs, fieldStruct)
						return
					}
					var noExportedFields = len(fieldStruct.Map()) == 0
					if noExportedFields {
						addIfNotExists(annotations, getName(field), getValue(field))
						return
					}
					var subAnnotations = make(map[string]interface{})
					addIfNotExists(annotations, getName(field), subAnnotations)
					buildAnnotations(fieldStruct, subAnnotations)
				} else {
					addIfNotExists(annotations, getName(field), getValue(field))
				}
			}(field, annotations)
		}
		wg.Wait()
		strucs = strucs[1:]
	}
}

func addIfNotExists(m map[string]interface{}, key string, value interface{}) {
	mutex.Lock()
	if _, ok := m[key]; !ok {
		m[key] = value
	}
	mutex.Unlock()
}
