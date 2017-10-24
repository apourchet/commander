package commander

import (
	"flag"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Commander is the struct that CLI applications will interact with
// to run their code.
type Commander struct{}

// New creates a new instance of the Commander.
func New() Commander { return Commander{} }

// RunCLI runs an application given with the command line arguments specified.
func (commander Commander) RunCLI(app interface{}, arguments []string) error {
	// Get the flagset from the tags of the app struct
	flagset, err := commander.GetFlagSet(app)
	if err != nil || flagset == nil {
		return errors.Wrap(err, "Failed to get flagset")
	}

	// Parse the arguments into that flagset
	flagset.Parse(arguments)

	// Execute the first argument
	args := flag.Args()
	if len(args) == 0 {
		return errors.Wrap(commander.PrintUsage(app), "Failed to print usage information")
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
		return errors.Wrap(commander.PrintUsage(app), "Failed to print usage information")
	}

	// Finally run that command if everything seems fine
	err = commander.RunCommand(app, args[0], args[1:]...)
	return errors.Wrapf(err, "Failed to run command %v", args[0])
}

// RunCommand runs a specific command of the application with arguments.
func (commander Commander) RunCommand(app interface{}, cmd string, args ...string) error {
	return nil
}

// HasCommand returns true if the application implements a specific command; and false otherwise.
func (commander Commander) HasCommand(app interface{}, cmd string) (bool, error) {
	return false, nil
}

// SubCommand returns the subcommand struct that corresponds to the command cmd. If none is found,
// SubCommand returns nil, nil.
func (commander Commander) SubCommand(app interface{}, cmd string) (interface{}, error) {
	return nil, nil
}

// GetFlagSet returns a flagset that corresponds to an application. This does not get
// return a flagset that will work for subcommands of that application.
func (commander Commander) GetFlagSet(app interface{}) (*flag.FlagSet, error) {
	st, valid := DerefStruct(app)
	if !valid {
		return nil, fmt.Errorf("An application needs to be a struct or a pointer to a struct")
	}

	flagset := flag.NewFlagSet("commander-main", flag.PanicOnError)
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup("commander"); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 {
				return nil, fmt.Errorf("Malformed tag on application: %v", alias)
			}
			if split[0] != "flag" {
				continue
			}
			err := SetFlag(flagset, app, field, split[1])
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to setup flag for application")
			}
		}
	}
	return flagset, nil
}

// Usage returns the "help" string for this application.
func (commander Commander) Usage(app interface{}) (string, error) {
	return "", nil
}

// PrintUsage prints the usage of the application given to the io.Writer specified; unless the
// Commander fails to get the usage for this application.
func (commander Commander) PrintUsage(app interface{}) error {
	usage, err := commander.Usage(app)
	if err != nil {
		return errors.Wrap(err, "Failed to generate usage text")
	}
	fmt.Println(usage)
	return nil
}
