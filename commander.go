package commander

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Commander is the struct that CLI applications will interact with
// to run their code.
type Commander struct {
	UsageOutput io.Writer
}

// New creates a new instance of the Commander.
func New() Commander {
	return Commander{
		UsageOutput: os.Stdout,
	}
}

// RunCLI runs an application given with the command line arguments specified.
func (commander Commander) RunCLI(app interface{}, arguments []string) error {
	// Get the flagset from the tags of the app struct
	flagset, err := commander.GetFlagSet(app)
	if err != nil || flagset == nil {
		return errors.Wrap(err, "Failed to get flagset")
	}

	// Parse the arguments into that flagset
	flagset.Parse(arguments)
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
		return errors.Wrapf(err, "Failed to search for subcommand %v", args[0])
	} else if subapp != nil {
		return commander.RunCLI(subapp, args[1:])
	}

	// Then check if there is a command with this name, and exit if there are errors
	if found, err := commander.HasCommand(app, args[0]); err != nil {
		return errors.Wrapf(err, "Failed to search for command %v", args[0])
	} else if !found {
		commander.PrintUsage(app)
		return fmt.Errorf("Failed to find command %v", args[0])
	}

	// Finally run that command if everything seems fine
	err := commander.RunCommand(app, args[0], args[1:]...)
	return errors.Wrapf(err, "Failed to run command %v", args[0])
}

// RunCommand runs a specific command of the application with arguments.
func (commander Commander) RunCommand(app interface{}, cmd string, args ...string) error {
	// TODO
	return nil
}

// SubCommand returns the subcommand struct that corresponds to the command cmd. If none is found,
// SubCommand returns nil, nil.
func (commander Commander) SubCommand(app interface{}, cmd string) (interface{}, error) {
	// TODO
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
	flagset := flag.NewFlagSet("commander-main", flag.ExitOnError)
	err := commander.SetupFlagSet(app, flagset)
	return flagset, errors.Wrapf(err, "Failed to get flagset")
}

// SetupFlagSet goes through the type of the application and creates flags on the flagset passed in.
func (commander Commander) SetupFlagSet(app interface{}, flagset *flag.FlagSet) error {
	// Get the raw type of the app
	st, valid := DerefStruct(app)
	if !valid {
		return fmt.Errorf("An application needs to be a struct or a pointer to a struct")
	}

	// Look through each field for flags and subcommand flags
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup("commander"); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 {
				return fmt.Errorf("Malformed tag on application: %v", alias)
			}

			// If this field is itself a flag
			if split[0] == "flag" {
				err := SetFlag(app, flagset, field, split[1])
				if err != nil {
					return errors.Wrapf(err, "Failed to setup flag for application")
				}
			}

			// If this field has subflags, recurse inside that
			if split[0] == "subcommand" {
				v, valid := Deref(app)
				if !valid || v.Kind() != reflect.Struct {
					return fmt.Errorf("Failed to get flags from field %v of type %v", field.Name, st.Name())
				}
				fieldval := v.FieldByName(field.Name)
				if !fieldval.IsValid() {
					return fmt.Errorf("Failed to get flags from field %v of type %v", field.Name, st.Name())
				}
				if err := commander.SetupFlagSet(fieldval.Interface(), flagset); err != nil {
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
	fmt.Fprintf(&buf, "\nSub-Commands:\n")
	st, valid := DerefStruct(app)
	if valid {
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			if alias, ok := field.Tag.Lookup("commander"); ok && alias != "" {
				split := strings.Split(alias, "=")
				if len(split) != 2 {
					continue
				}

				// If this field has subflags, recurse inside that
				if split[0] == "subcommand" {
					inner := strings.Split(split[1], ",")
					line := strings.Join(inner, "  |  ")
					fmt.Fprintf(&buf, "  %v\n", line)
				}
			}
		}
	}
	fmt.Fprintf(&buf, "\n")
	return buf.String()
}

// PrintUsage prints the usage of the application given to the io.Writer specified; unless the
// Commander fails to get the usage for this application.
func (commander Commander) PrintUsage(app interface{}) {
	usage := commander.Usage(app)
	fmt.Fprintf(commander.UsageOutput, usage)
}
