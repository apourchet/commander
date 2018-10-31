package commander

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/apourchet/commander/utils"
	"github.com/pkg/errors"
)

const (
	// FieldTag is the name of the field tag that commander uses
	FieldTag = "commander"

	// DefaultCommand is the default name of the method that will be called on the application
	// objects.
	DefaultCommand = "CommanderDefault"

	// SubcommandDirective indicates a subcommand
	SubcommandDirective = "subcommand"

	// FlagStructDirective indicates that the field is a struct containing flags to
	// inject. Commander will go into that struct and populate its fields if they
	// are tagged with a FlagDirective.
	FlagStructDirective = "flagstruct"

	// FlagSliceDirective indicates that the field is a slice containing structs
	// that need flags to be injected into. Commander will go through each struct
	// in the slice as though it had a FlagStruct directive.
	FlagSliceDirective = "flagslice"

	// FlagDirective indicates that this field should be populated using the command
	// line flags
	FlagDirective = "flag"
)

// NamedCLI is the interface that the application should implement to change the default displayed
// name when Usage is called.
type NamedCLI interface {
	CLIName() string
}

// PostFlagParseHook is the interface that the application should implement to receive a callback
// when the flags have been injected into it.
type PostFlagParseHook interface {
	PostFlagParse() error
}

// Commander is the struct that CLI applications will interact with
// to run their code.
type Commander struct {
	UsageOutput       io.Writer
	FlagErrorHandling flag.ErrorHandling
}

// New creates a new instance of the Commander.
func New() Commander {
	return Commander{
		UsageOutput:       os.Stdout,
		FlagErrorHandling: flag.ExitOnError,
	}
}

// RunCLI runs an application given with the command line arguments specified.
func (commander Commander) RunCLI(app interface{}, arguments []string) error {
	// Get the flagset from the tags of the app struct
	flagset, err := commander.GetFlagSet(app)
	if err != nil {
		return errors.WithStack(err)
	}

	return commander.RunCLIWithFlagSet(app, arguments, flagset)
}

// RunCLIWithFlagSet runs the cli with the flagset passed in. This is useful for ad-hoc flags that
// are not bound to fields within the application.
func (commander Commander) RunCLIWithFlagSet(app interface{}, args []string, flagset *flag.FlagSet) error {
	// Parse the arguments into that flagset
	err := flagset.Parse(args)
	if err != nil {
		return errors.WithStack(err)
	}

	// Execute the first argument
	args = flagset.Args()
	if len(args) == 0 {
		args = []string{DefaultCommand}
	}
	cmd := args[0]

	// Check first if there is a subcommand with this name
	if subapp, err := commander.SubCommand(app, cmd); err != nil {
		commander.PrintUsage(app)
		return errors.Wrapf(err, "Failed to search for subcommand %v", cmd)
	} else if subapp != nil {
		if err = executeHook(app); err != nil {
			return errors.WithStack(err)
		}
		return commander.RunCLI(subapp, args[1:])
	}

	// Then check if there is a command with this name, and exit if there are errors
	if found, err := commander.HasCommand(app, cmd); err != nil {
		commander.PrintUsage(app)
		return errors.Wrapf(err, "Failed to search for command %v", cmd)
	} else if !found {
		if foundDefault, err := commander.HasCommand(app, DefaultCommand); err != nil {
			commander.PrintUsage(app)
			return errors.Wrapf(err, "Failed to search for command %v", cmd)
		} else if !foundDefault {
			commander.PrintUsage(app)
			if cmd != DefaultCommand {
				return fmt.Errorf("Failed to find command %v or %v", cmd, DefaultCommand)
			}
			return fmt.Errorf("Failed to find default command %v", DefaultCommand)
		} else {
			cmd = DefaultCommand
			args = append([]string{DefaultCommand}, args...)
		}
	}

	// Reparse flags to populate some of the flags that the default package might have
	// missed
	err = flagset.Parse(args[1:])
	if err != nil {
		return errors.WithStack(err)
	}
	args = flagset.Args()

	// Execute post flag parse hook
	if err = executeHook(app); err != nil {
		return errors.WithStack(err)
	}

	// Finally run that command if everything seems fine
	err = commander.RunCommand(app, cmd, args...)
	if err != nil {
		commander.PrintUsage(app)
		return errors.Wrapf(err, "Failed to run command %v", cmd)
	}
	return nil
}

// RunCommand runs a specific command of the application with arguments.
func (commander Commander) RunCommand(app interface{}, cmd string, args ...string) error {
	apptype := reflect.TypeOf(app)
	for i := 0; i < apptype.NumMethod(); i++ {
		// Find the right method
		method := apptype.Method(i)
		if strings.ToLower(method.Name) != strings.ToLower(cmd) {
			continue
		}

		// Make sure we have enough args for this command
		inputsize := method.Type.NumIn() - 1
		if len(args) != inputsize && method.Type.In(inputsize).Kind() != reflect.Slice {
			return fmt.Errorf("Command %v requires %v arguments, have %v", cmd, inputsize, len(args))
		} else if len(args) < inputsize {
			args = append(args, "[]")
		} else if len(args) > inputsize || method.Type.In(inputsize).Kind() == reflect.Slice {
			// Then we consider that the extra arguments are just a list of strings
			extras := args[inputsize-1:]
			bytes, _ := json.Marshal(extras)
			args[inputsize-1] = string(bytes)
			args = args[:inputsize]
		}

		in := make([]reflect.Value, len(args)+1)
		in[0] = reflect.ValueOf(app)
		for i, arg := range args {
			t := method.Type.In(i + 1)
			param, err := utils.ParseString(t, arg)
			if err != nil {
				return errors.Wrapf(err, "Failed to parse string into function argument")
			}
			in[i+1] = param
		}
		method.Func.Call(in)
		return nil
	}
	return fmt.Errorf("Failed to find method %v", cmd)
}

// SubCommand returns the subcommand struct that corresponds to the command cmd. If none is found,
// SubCommand returns nil, nil.
func (commander Commander) SubCommand(app interface{}, cmd string) (interface{}, error) {
	st, valid := utils.DerefType(app)
	if !valid {
		return nil, fmt.Errorf("An application needs to be a struct or a pointer to a struct")
	}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 && (split[0] == FlagDirective || split[0] == SubcommandDirective) {
				return nil, fmt.Errorf("Malformed tag on application: %v", alias)
			}

			// If this field has subflags, recurse inside that
			if split[0] != SubcommandDirective {
				continue
			}

			// Parse the directive to get the subcommand
			subcmd, _ := parseSubcommandDirective(split[1])
			if subcmd != cmd {
				continue
			}

			// We have found the right subcommand
			v, valid := utils.DerefValue(app)
			if !valid || v.Kind() != reflect.Struct {
				return nil, fmt.Errorf("Failed to get subcommand from field %v of type %v", field.Name, st.Name())
			}
			fieldval := v.FieldByName(field.Name)
			if !fieldval.IsValid() {
				return nil, fmt.Errorf("Failed to get subcommand from field %v of type %v", field.Name, st.Name())
			}
			return fieldval.Interface(), nil
		}
	}
	return nil, nil
}

// HasCommand returns true if the application implements a specific command; and false otherwise.
func (commander Commander) HasCommand(app interface{}, cmd string) (bool, error) {
	cmd = strings.Replace(cmd, "-", "", -1)
	cmd = strings.Replace(cmd, "_", "", -1)
	cmd = strings.ToLower(cmd)
	apptype := reflect.TypeOf(app)
	for i := 0; i < apptype.NumMethod(); i++ {
		method := apptype.Method(i)
		if strings.ToLower(method.Name) == cmd {
			return true, nil
		}
	}
	return false, nil
}

// GetFlagSet returns a flagset that corresponds to an application. This does not get
// return a flagset that will work for subcommands of that application.
func (commander Commander) GetFlagSet(app interface{}) (*flag.FlagSet, error) {
	appname := "CLI"
	if casted, ok := app.(NamedCLI); ok {
		appname = casted.CLIName()
	}

	flagset := flag.NewFlagSet(appname, commander.FlagErrorHandling)
	setter := newFlagSetter(flagset)
	defer setter.finish()

	err := commander.setupFlagSet(app, setter)
	return flagset, errors.Wrapf(err, "Failed to get flagset")
}

// SetupflagSet goes through the type of the application and creates flags on the flagset passed in.
func (commander Commander) setupFlagSet(app interface{}, setter *flagSetter) error {
	// Get the raw type of the app
	st, valid := utils.DerefType(app)
	if !valid {
		return fmt.Errorf("An application needs to be a struct or a pointer to a struct")
	}

	// Look through each field for flags and subcommand flags
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 && (split[0] == FlagDirective || split[0] == SubcommandDirective) {
				return fmt.Errorf("Malformed tag on application: %v", alias)
			}

			// If this field is itself a flag
			if split[0] == FlagDirective {
				err := setter.setFlag(app, field, split[1])
				if err != nil {
					return errors.Wrapf(err, "Failed to setup flag for application")
				}
			}

			// If this field has subflags, recurse inside that
			if split[0] == FlagStructDirective {
				v, valid := utils.DerefValue(app)
				if !valid || v.Kind() != reflect.Struct {
					// The subapp is nil or not a struct
					return nil
				}
				fieldval := v.FieldByName(field.Name)
				if !fieldval.IsValid() {
					return fmt.Errorf("Failed to get flags from field %v of type %v", field.Name, st.Name())
				}
				fieldIface := fieldval.Interface()
				if fieldval.Type().Kind() == reflect.Struct {
					fieldIface = fieldval.Addr().Interface()
				}
				if err := commander.setupFlagSet(fieldIface, setter); err != nil {
					return errors.Wrap(err, "Failed to get flagset for sub-struct")
				}
			} else if split[0] == FlagSliceDirective {
				v, valid := utils.DerefValue(app)
				if !valid || v.Kind() != reflect.Struct {
					// The subapp is nil or not a struct
					return nil
				}
				fieldval := v.FieldByName(field.Name)
				if !fieldval.IsValid() {
					return fmt.Errorf("Failed to get flags from field %v of type %v", field.Name, st.Name())
				} else if fieldval.Kind() != reflect.Slice {
					return fmt.Errorf("FlagSlice directive should only be used on slice fields")
				}
				for i := 0; i < fieldval.Len(); i++ {
					item := fieldval.Index(i)
					if err := commander.setupFlagSet(item.Interface(), setter); err != nil {
						return errors.Wrap(err, "Failed to get flagset for slice element")
					}
				}
			}
		}
	}
	return nil
}

// Usage returns the "help" string for this application.
func (commander Commander) Usage(app interface{}) string {
	var buf bytes.Buffer

	// First use the flagset to print flags
	flagset, err := commander.GetFlagSet(app)
	if err == nil && flagset != nil {
		flagset.SetOutput(&buf)
		flagset.Usage()
	}

	// Then print subcommands
	st, valid := utils.DerefType(app)
	if !valid {
		return buf.String()
	}

	directives := []string{}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 || split[0] != SubcommandDirective {
				continue
			}

			directives = append(directives, split[1])
		}
	}

	if len(directives) == 0 {
		return buf.String()
	}

	fmt.Fprintf(&buf, "\nSub-Commands:\n")
	for _, directive := range directives {
		// If this field has subflags, recurse inside that
		cmd, desc := parseSubcommandDirective(directive)
		fmt.Fprintf(&buf, "  %v  |  %v\n", cmd, desc)
	}

	return buf.String()
}

// PrintUsage prints the usage of the application given to the io.Writer specified; unless the
// Commander fails to get the usage for this application.
func (commander Commander) PrintUsage(app interface{}) {
	usage := commander.Usage(app)
	fmt.Fprintf(commander.UsageOutput, usage)
}

// ParseSubcommandDirective parses the subcommand directive into the subcommand string and its description.
func parseSubcommandDirective(directive string) (cmd string, description string) {
	split := strings.SplitN(directive, ",", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return split[0], "No description for this subcommand"
}

func executeHook(app interface{}) error {
	if hook, ok := app.(PostFlagParseHook); ok {
		if err := hook.PostFlagParse(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
