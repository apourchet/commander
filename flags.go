package commander

import (
	"flag"
	"reflect"
	"strings"

	"github.com/apourchet/commander/utils"
)

// FlagTarget are the structs that the std::flag package will interact with. FlagTargets
// will populate the values of the fields of the given object through the Set function
// that the std::flag package calls when a flag is defined.
type FlagTarget struct {
	object interface{}
	field  reflect.StructField
}

// NewFlagTarget creates a new FlagTarget that points to the object given.
func NewFlagTarget(object interface{}, field reflect.StructField) *FlagTarget {
	flagtarget := &FlagTarget{
		object: object,
		field:  field,
	}
	return flagtarget
}

// String returns the stringified value of the object's field that the FlagTarget is bound to.
func (flagtarget *FlagTarget) String() string {
	// TODO: return default value
	return " "
}

// Set sets the value of the field that the FlagTarget is bound to.
func (flagtarget *FlagTarget) Set(value string) error {
	return utils.SetField(flagtarget.object, flagtarget.field.Name, value)
}

// SetFlag creates a flag on the flagset given so that when the flagset.
func SetFlag(obj interface{}, flagset *flag.FlagSet, field reflect.StructField, directive string) error {
	name, usage := ParseFlagDirective(directive)

	if field.Type.Kind() == reflect.Bool {
		// Bool is special because it doesn't need a value
		v := reflect.ValueOf(obj)
		var ptr *bool
		if v.Kind() == reflect.Ptr {
			v = v.Elem().FieldByName(field.Name)
			ptr = v.Addr().Interface().(*bool)
			flagset.BoolVar(ptr, name, false, usage) // TODO: default value
			return nil
		}
	}

	flagtarget := NewFlagTarget(obj, field)
	flagset.Var(flagtarget, name, usage)
	return nil
}

// ParseFlagDirective parses the directive into the flag's name and its usage. The format of a flag directive is
// <name>,<usage>.
func ParseFlagDirective(directive string) (name string, usage string) {
	split := strings.SplitN(directive, ",", 2)
	if len(split) == 1 {
		return directive, "No usage found for this flag."
	}
	return split[0], split[1]
}
