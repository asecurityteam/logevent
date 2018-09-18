package logevent

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structs"
)

const (
	tagKey       = "logevent"
	unknown      = "unknown"
	defaultValue = "default="
)

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

func buildAnnotations(s *structs.Struct, annotations map[string]interface{}) {
	s.TagName = tagKey
	var strucs = []*structs.Struct{s}
	for len(strucs) > 0 {
		for _, field := range strucs[0].Fields() {
			if structs.IsStruct(field.Value()) && field.IsEmbedded() {
				strucs = append(strucs, structs.New(field.Value()))
				continue
			}
			if structs.IsStruct(field.Value()) {
				var subAnnotations = make(map[string]interface{})
				addIfNotExists(annotations, getName(field), subAnnotations)
				buildAnnotations(structs.New(field.Value()), subAnnotations)
			} else {
				addIfNotExists(annotations, getName(field), getValue(field))
			}
		}
		strucs = strucs[1:]
	}
}

func addIfNotExists(m map[string]interface{}, key string, value interface{}) {
	if _, ok := m[key]; !ok {
		m[key] = value
	}
}
