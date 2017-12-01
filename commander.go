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

	// SubcommandDirective indicates a subcommand
	SubcommandDirective = "subcommand"

	// FlagStructDirective indicates that the field
	// is a struct containing flags to inject
	FlagStructDirective = "flagstruct"

	// FlagDirective indicates that this field should be populated using
	// the command line flags
	FlagDirective = "flag"
)

type namedCLI interface {
	CLIName() string
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
		FlagErrorHandling: flag.ContinueOnError,
	}
}

// RunCLI runs an application given with the command line arguments specified.
func (commander Commander) RunCLI(app interface{}, arguments []string) error {
	// Get the flagset from the tags of the app struct
	flagset, err := commander.GetFlagSet(app)
	if err != nil {
		return errors.WithStack(err)
	}

	// Parse the arguments into that flagset
	err = flagset.Parse(arguments)
	if err != nil {
		return errors.WithStack(err)
	}
	return commander.RunCLIWithFlagSet(app, arguments, flagset)
}

// RunCLIWithFlagSet runs the cli with the flagset passed in. This is useful for ad-hoc flags that
// are not bound to fields within the application.
func (commander Commander) RunCLIWithFlagSet(app interface{}, arguments []string, flagset *flag.FlagSet) error {
	// Execute the first argument
	args := flagset.Args()
	if len(args) == 0 {
		args = []string{"basecommand"}
	}

	// Check first if there is a subcommand with this name
	if subapp, err := commander.SubCommand(app, args[0]); err != nil {
		commander.PrintUsage(app)
		return errors.Wrapf(err, "Failed to search for subcommand %v", args[0])
	} else if subapp != nil {
		return commander.RunCLI(subapp, args[1:])
	}

	// Then check if there is a command with this name, and exit if there are errors
	if found, err := commander.HasCommand(app, args[0]); err != nil {
		commander.PrintUsage(app)
		return errors.Wrapf(err, "Failed to search for command %v", args[0])
	} else if !found {
		commander.PrintUsage(app)
		return fmt.Errorf("Failed to find command %v", args[0])
	}

	// Finally run that command if everything seems fine
	err := commander.RunCommand(app, args[0], args[1:]...)
	if err != nil {
		commander.PrintUsage(app)
		return errors.Wrapf(err, "Failed to run command %v", args[0])
	}
	return nil
}

// RunCommand runs a specific command of the application with arguments.
func (commander Commander) RunCommand(app interface{}, cmd string, args ...string) error {
	apptype := reflect.TypeOf(app)
	for i := 0; i < apptype.NumMethod(); i++ {
		// Find the right method
		method := apptype.Method(i)
		if strings.ToLower(method.Name) != cmd {
			continue
		}

		// Make sure we have enough args for this command
		inputsize := method.Type.NumIn()
		if len(args) < inputsize-1 && method.Type.In(inputsize-1).Kind() != reflect.Slice {
			return fmt.Errorf("Command %v requires %v arguments, have %v", cmd, inputsize-1, len(args))
		} else if len(args) < inputsize-1 {
			args = append(args, "[]")
		} else if len(args) > inputsize-1 || method.Type.In(inputsize-1).Kind() == reflect.Slice {
			// Then we consider that the extra arguments are just a list of strings
			extras := args[inputsize-2:]
			bytes, _ := json.Marshal(extras)
			args[inputsize-2] = string(bytes)
			args = args[:inputsize-1]
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
			subcmd, _ := ParseSubcommandDirective(split[1])
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
	appname := "commander-cli"
	if casted, ok := app.(namedCLI); ok {
		appname = casted.CLIName()
	}
	flagset := flag.NewFlagSet(appname, commander.FlagErrorHandling)
	err := commander.SetuflagSet(app, flagset)
	return flagset, errors.Wrapf(err, "Failed to get flagset")
}

// SetuflagSet goes through the type of the application and creates flags on the flagset passed in.
func (commander Commander) SetuflagSet(app interface{}, flagset *flag.FlagSet) error {
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
				err := SetFlag(app, flagset, field, split[1])
				if err != nil {
					return errors.Wrapf(err, "Failed to setup flag for application")
				}
			}

			// If this field has subflags, recurse inside that
			if split[0] == SubcommandDirective || split[0] == FlagStructDirective {
				v, valid := utils.DerefValue(app)
				if !valid || v.Kind() != reflect.Struct {
					// The subapp is nil or not a struct
					return nil
				}
				fieldval := v.FieldByName(field.Name)
				if !fieldval.IsValid() {
					return fmt.Errorf("Failed to get flags from field %v of type %v", field.Name, st.Name())
				}
				if err := commander.SetuflagSet(fieldval.Interface(), flagset); err != nil {
					return errors.Wrap(err, "Failed to get flagset for sub-struct")
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
	if valid && st.NumField() > 0 {
		fmt.Fprintf(&buf, "\nSub-Commands:\n")
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
				split := strings.Split(alias, "=")
				if len(split) != 2 || split[0] != SubcommandDirective {
					continue
				}

				// If this field has subflags, recurse inside that
				cmd, desc := ParseSubcommandDirective(split[1])
				fmt.Fprintf(&buf, "  %v  |  %v\n", cmd, desc)
			}
		}
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
func ParseSubcommandDirective(directive string) (cmd string, description string) {
	split := strings.SplitN(directive, ",", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return split[0], "No description for this subcommand"
}
