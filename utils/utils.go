package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// DerefType dereferences the type until it is of kind struct.
func DerefType(obj interface{}) (reflect.Type, bool) {
	st := reflect.TypeOf(obj)
	for st.Kind() == reflect.Ptr || st.Kind() == reflect.Interface {
		st = st.Elem()
	}
	return st, st.Kind() == reflect.Struct
}

// DerefValue dereferences pointers and interfaces until they become their base types.
// Returns false if the inner interface is nil after dereferencing.
func DerefValue(obj interface{}) (reflect.Value, bool) {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() { // If the chain ends in a nil, skip this
			return v, false
		}
		v = v.Elem()
	}
	return v, true
}

// Stringify returns the string representation of the value given. It functions like fmt.Printf("%v")
// except for slices and maps; where it json stringifies them.
func Stringify(val interface{}) (string, error) {
	v, valid := DerefValue(val)
	if !valid {
		return "", nil
	}
	return StringifyValue(reflect.ValueOf(v.Interface()))
}

// StringifyValue returns the string representation of the value given. It functions like fmt.Printf("%v")
// except for slices and maps; where it json stringifies them.
func StringifyValue(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Ptr:
		return StringifyValue(v.Elem())
	case reflect.Bool:
		return fmt.Sprintf("%v", v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", v.Float()), nil
	case reflect.String:
		return fmt.Sprintf("%v", v.String()), nil
	case reflect.Slice, reflect.Map:
		val := v.Interface()
		content, err := json.Marshal(val)
		if err != nil {
			return "", fmt.Errorf("Failed to stringify value into url: %v", err)
		}
		return string(content), nil
	}
	return "", fmt.Errorf("Unsupported type: %v", v.Kind())
}

// GetFieldValue returns the stringified value of the field by name given the object.
func GetFieldValue(obj interface{}, fieldname string) (string, error) {
	v, valid := DerefValue(obj)
	if !valid || v.Kind() != reflect.Struct {
		return "", nil
	}

	field := v.FieldByName(fieldname)
	if !field.IsValid() {
		return "", fmt.Errorf("Field not found when setting field: %s", fieldname)
	}

	str, err := StringifyValue(field)
	return str, err
}

// SetField sets the field of the object using a string that
// was retrieved from the URI of the request.
func SetField(obj interface{}, fieldname, value string) error {
	v, valid := DerefValue(obj)
	if !valid || v.Kind() != reflect.Struct {
		return nil
	}

	field := v.FieldByName(fieldname)
	if !field.IsValid() {
		return fmt.Errorf("Field not found when setting field: %s", fieldname)
	}

	val, err := ParseString(field.Type(), value)
	if err != nil {
		return fmt.Errorf("Failed to parse value: %v", err)
	}

	field.Set(val)
	return nil
}

// ParseString parses the string into a value depending on the type that gets passed in.
// time.Duration is handled separately because of the fact that its an int64 with some fancy parsing involved.
func ParseString(t reflect.Type, value string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Ptr:
		subval, err := ParseString(t.Elem(), value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %s: %v", t, err)
		}
		val := reflect.New(t.Elem())
		val.Elem().Set(subval)
		return val, nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", b, err)
		}
		return reflect.ValueOf(b), nil
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(int(i)), nil
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(int8(i)), nil
	case reflect.Int16:
		i, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(int16(i)), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(int32(i)), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return reflect.ValueOf(int64(i)), nil
		}
		dur, err := time.ParseDuration(value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T or %T: %v", i, dur, err)
		}
		return reflect.ValueOf(dur), nil
	case reflect.Uint:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(uint(i)), nil
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(uint8(i)), nil
	case reflect.Uint16:
		i, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(uint16(i)), nil
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(uint32(i)), nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", i, err)
		}
		return reflect.ValueOf(uint64(i)), nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", f, err)
		}
		return reflect.ValueOf(float32(f)), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", f, err)
		}
		return reflect.ValueOf(float64(f)), nil
	case reflect.Slice:
		s := []string{}
		err := json.Unmarshal([]byte(value), &s)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", s, err)
		}
		return reflect.ValueOf(s), nil
	case reflect.Map:
		m := map[string]string{}
		err := json.Unmarshal([]byte(value), &m)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse string to %T: %v", m, err)
		}
		return reflect.ValueOf(m), nil
	}
	return reflect.ValueOf(nil), fmt.Errorf("Unsupported type: %v", t)
}
