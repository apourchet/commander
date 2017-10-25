package commander

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// DerefStruct dereferences the type until it is of kind struct.
func DerefStruct(obj interface{}) (reflect.Type, bool) {
	st := reflect.TypeOf(obj)
	for st.Kind() == reflect.Ptr || st.Kind() == reflect.Interface {
		st = st.Elem()
	}
	return st, st.Kind() == reflect.Struct
}

// Deref dereferences pointers and interfaces until they become their base types.
func Deref(obj interface{}) (reflect.Value, bool) {
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
func Stringify(val interface{}) (bool, string, error) {
	v, valid := Deref(val)
	if !valid {
		return true, "", nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return false, fmt.Sprintf("%v", v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return false, fmt.Sprintf("%v", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		return false, fmt.Sprintf("%v", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return false, fmt.Sprintf("%v", v.Float()), nil
	case reflect.String:
		return false, fmt.Sprintf("%v", v.String()), nil
	case reflect.Slice, reflect.Map:
		content, err := json.Marshal(val)
		if err != nil {
			return false, "", fmt.Errorf("Failed to stringify value into url: %v", err)
		}
		return false, string(content), nil
	}
	return false, "", fmt.Errorf("Unsupported type: %T", val)
}

// SetField sets the field of the object using a string that
// was retrieved from the URI of the request.
func SetField(obj interface{}, fieldname, value string) error {
	v, valid := Deref(obj)
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
func ParseString(t reflect.Type, value string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Ptr:
		subval, err := ParseString(t.Elem(), value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %s: %v", t, err)
		}
		val := reflect.New(t.Elem())
		val.Elem().Set(subval)
		return val, nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", b, err)
		}
		return reflect.ValueOf(b), nil
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(int(i)), nil
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(int8(i)), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(int32(i)), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(int64(i)), nil
	case reflect.Uint:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(uint(i)), nil
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(uint8(i)), nil
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(uint32(i)), nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", i, err)
		}
		return reflect.ValueOf(uint64(i)), nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", f, err)
		}
		return reflect.ValueOf(float32(f)), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", f, err)
		}
		return reflect.ValueOf(float64(f)), nil
	case reflect.Slice:
		var s []interface{}
		err := json.Unmarshal([]byte(value), &s)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", s, err)
		}
		return reflect.ValueOf(s), nil
	case reflect.Map:
		var m map[string]interface{}
		err := json.Unmarshal([]byte(value), &m)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse flag to %T: %v", m, err)
		}
		return reflect.ValueOf(m), nil
	}
	return reflect.ValueOf(nil), fmt.Errorf("Unsupported type: %v", t)
}

// ParseFlagDirective parses the directive into the flag's name and its usage. The format of a flag directive is
// <name>,<usage>
func ParseFlagDirective(directive string) (name string, usage string) {
	split := strings.SplitN(directive, ",", 2)
	if len(split) == 1 {
		return directive, "No usage found for this flag."
	}
	return split[0], split[1]
}