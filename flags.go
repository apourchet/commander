package commander

import (
	"flag"
	"reflect"
)

// FlagTarget are the structs that the std::flag package will interact with. FlagTargets
// will populate the values of the fields of the given object through the Set function
// that the std::flag package calls when a flag is defined.
type FlagTarget struct {
	object interface{}
	field  reflect.StructField
}

// NewFlagTarget creates a new FlagTarget that points to the object given.
func NewFlagTarget(object interface{}, field reflect.StructField) FlagTarget {
	flagtarget := FlagTarget{
		object: object,
		field:  field,
	}
	return flagtarget
}

// String returns the stringified value of the object's field that the FlagTarget is bound to.
func (flagtarget FlagTarget) String() string {
	v, valid := Deref(flagtarget.object)
	if !valid || v.Kind() != reflect.Struct {
		return "Failed to stringify field"
	}

	field := v.FieldByName(flagtarget.field.Name)
	if !field.IsValid() {
		return "Failed to stringify field"
	}

	ok, str, err := Stringify(field)
	if err != nil || !ok {
		return "Failed to stringify field"
	}
	return str
}

// Set sets the value of the field that the FlagTarget is bound to.
func (flagtarget FlagTarget) Set(value string) error {
	return SetField(flagtarget.object, flagtarget.field.Name, value)
}

// SetFlag creates a flag on the flagset given so that when the flagset.
func SetFlag(obj interface{}, flagset *flag.FlagSet, field reflect.StructField, directive string) error {
	name, usage := ParseFlagDirective(directive)

	// TODO: Bool is special because it doesn't need a value.
	// if field.Type.Kind() == reflect.Bool {
	// 	v := reflect.ValueOf(obj)
	// 	var ptr *bool
	// 	if v.Kind() == reflect.Ptr {
	// 		v = v.Elem().FieldByName(field.Name)
	// 		ptr = v.Addr().Interface().(*bool)
	// 		flagset.BoolVar(ptr, name, false, usage)
	// 		return nil
	// 	}
	// }

	flagtarget := NewFlagTarget(obj, field)
	flagset.Var(flagtarget, name, usage)
	return nil
}
