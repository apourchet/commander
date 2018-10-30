package commander

import (
	"flag"
	"fmt"
	"reflect"
	"strings"

	"github.com/apourchet/commander/utils"
	"github.com/pkg/errors"
)

// flagTarget are the structs that the std::flag package will interact with. FlagTargets
// will populate the values of the fields of the given object through the Set function
// that the std::flag package calls when a flag is defined.
type flagTarget struct {
	object interface{}
	field  reflect.StructField
	usage  string
}

// newFlagTarget creates a new FlagTarget that points to the object given.
func newFlagTarget(obj interface{}, field reflect.StructField, usage string) *flagTarget {
	flagtarget := &flagTarget{
		object: obj,
		field:  field,
		usage:  usage,
	}
	return flagtarget
}

func (target *flagTarget) Usage() string {
	def, _ := utils.GetFieldValue(target.object, target.field.Name)
	if target.field.Type.Kind() == reflect.String {
		def = fmt.Sprintf(`"%s"`, def)
	}
	return fmt.Sprintf(`%s (type: %s, default: %s)`, target.usage, target.field.Type.Kind(), def)
}

// String returns the stringified value of the object's field that the FlagTarget is bound to.
func (target *flagTarget) String() string {
	return ""
}

// IsBoolFlag returns false always so that the flag usage does not show "value" after each flag.
func (target *flagTarget) IsBoolFlag() bool {
	return target.field.Type.Kind() == reflect.Bool
}

// Set sets the value of the field that the FlagTarget is bound to.
func (target *flagTarget) Set(value string) error {
	if err := utils.SetField(target.object, target.field.Name, value); err != nil {
		return err
	}
	return nil
}

// FlagSetter is the wrapper around flag.FlagSet that allows setting of a flag multiple times. This is
// useful in the case of subcommands that might use the same flag.
type flagSetter struct {
	flagset *flag.FlagSet
	targets map[string]*flagTarget
}

// NewFlagSetter returns a new FlagSetter, with the internal variables initialized.
func newFlagSetter(flagset *flag.FlagSet) *flagSetter {
	return &flagSetter{
		flagset: flagset,
		targets: map[string]*flagTarget{},
	}
}

// SetFlag creates a flag on the flagset given so that when the flagset.
func (setter *flagSetter) setFlag(obj interface{}, field reflect.StructField, directive string) error {
	name, usage := parseFlagDirective(directive)
	return setter.addTarget(name, obj, field, usage)
}

// Finish tells the setter that the flags have all been accounted for, and it can forward all the flag
// setup to the internal flagset.
func (setter *flagSetter) finish() {
	for name, target := range setter.targets {
		setter.flagset.Var(target, name, target.Usage())
	}
}

func (setter *flagSetter) addTarget(name string, obj interface{}, field reflect.StructField, usage string) error {
	target, found := setter.targets[name]
	if found {
		return errors.Errorf("Duplicate binding of flag: %v", name)
	}
	target = newFlagTarget(obj, field, usage)
	setter.targets[name] = target
	return nil
}

// ParseFlagDirective parses the directive into the flag's name and its usage. The format of a flag directive is
// <name>,<usage>.
func parseFlagDirective(directive string) (name string, usage string) {
	split := strings.SplitN(directive, ",", 2)
	if len(split) == 1 {
		return directive, "No usage found for this flag."
	}
	return split[0], split[1]
}
