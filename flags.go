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
	objects []interface{}
	fields  []reflect.StructField
	usage   string
}

// NewFlagTarget creates a new FlagTarget that points to the object given.
func NewFlagTarget() *FlagTarget {
	flagtarget := &FlagTarget{
		objects: []interface{}{},
		fields:  []reflect.StructField{},
		usage:   "",
	}
	return flagtarget
}

func (target *FlagTarget) add(obj interface{}, field reflect.StructField, usage string) {
	target.objects = append(target.objects, obj)
	target.fields = append(target.fields, field)
	target.usage = usage
}

// String returns the stringified value of the object's field that the FlagTarget is bound to.
func (target *FlagTarget) String() string {
	// TODO: return default value
	return " "
}

// Set sets the value of the field that the FlagTarget is bound to.
func (target *FlagTarget) Set(value string) error {
	for i := 0; i < len(target.objects); i++ {
		if err := utils.SetField(target.objects[i], target.fields[i].Name, value); err != nil {
			return err
		}
	}
	return nil
}

// FlagSetter is the wrapper around flag.FlagSet that allows setting of a flag multiple times. This is
// useful in the case of subcommands that might use the same flag.
type FlagSetter struct {
	flagset *flag.FlagSet
	targets map[string]*FlagTarget
}

// NewFlagSetter returns a new FlagSetter, with the internal variables initialized.
func NewFlagSetter(flagset *flag.FlagSet) *FlagSetter {
	return &FlagSetter{
		flagset: flagset,
		targets: map[string]*FlagTarget{},
	}
}

// SetFlag creates a flag on the flagset given so that when the flagset.
func (setter *FlagSetter) SetFlag(obj interface{}, field reflect.StructField, directive string) error {
	name, usage := ParseFlagDirective(directive)

	if field.Type.Kind() == reflect.Bool {
		// Bool is special because it doesn't need a value
		v := reflect.ValueOf(obj)
		var ptr *bool
		if v.Kind() == reflect.Ptr {
			v = v.Elem().FieldByName(field.Name)
			ptr = v.Addr().Interface().(*bool)
			setter.flagset.BoolVar(ptr, name, false, usage) // TODO: default value
			return nil
		}
	}

	setter.addTarget(name, obj, field, usage)
	return nil
}

// Finish tells the setter that the flags have all been accounted for, and it can forward all the flag
// setup to the internal flagset.
func (setter *FlagSetter) Finish() {
	for name, target := range setter.targets {
		setter.flagset.Var(target, name, target.usage)
	}
}

func (setter *FlagSetter) addTarget(name string, obj interface{}, field reflect.StructField, usage string) {
	target, found := setter.targets[name]
	if !found {
		target = NewFlagTarget()
	}
	target.add(obj, field, usage)
	setter.targets[name] = target
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
