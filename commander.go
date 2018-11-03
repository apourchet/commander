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
	cumulativeCommands := []string{}
	originalApp := app
	appname := getCLIName(originalApp, cumulativeCommands...)
	for {
		// Get the flagset from the tags of the app struct
		flagset, err := commander.GetFlagSet(app, appname)
		if err != nil {
			return errors.WithStack(err)
		}

		// Parse the arguments into that flagset
		if err := flagset.Parse(arguments); err != nil {
			return errors.WithStack(err)
		}

		if arguments = flagset.Args(); len(arguments) > 0 {
			if subapp, err := subCommand(app, arguments[0]); err != nil {
				return errors.Wrapf(err, "failed to search for subcommand %v", arguments[0])
			} else if subapp != nil {
				if err = executeHook(app); err != nil {
					return errors.WithStack(err)
				}
				cumulativeCommands = append(cumulativeCommands, arguments[0])
				app = subapp
				arguments = arguments[1:]
				appname = getCLIName(originalApp, cumulativeCommands...)
				continue
			}
		}

		commands := []string{}
		if len(cumulativeCommands) > 0 {
			prevCmd := cumulativeCommands[len(cumulativeCommands)-1]
			commands = []string{prevCmd}
		}
		if len(arguments) > 0 {
			commands = append([]string{arguments[0]}, commands...)
			cumulativeCommands = append(cumulativeCommands, arguments[0])
		}
		commands = append(commands, DefaultCommand)

		cmd, err := findCommand(app, commands)
		if err != nil {
			return err
		} else if cmd == "" {
			commander.PrintUsage(app, appname)
			return fmt.Errorf("failed to find possible method: %v", commands)
		} else if len(arguments) > 0 && cmd == arguments[0] {
			if len(cumulativeCommands) < 2 || cumulativeCommands[len(cumulativeCommands)-2] != arguments[0] {
				arguments = arguments[1:]
			}
		}

		// TODO: setup additional flags with the command we are about to run

		err = executeCommand(app, cmd, arguments, flagset.FlagSet)
		if err != nil && !isApplicationError(err) {
			commander.PrintUsage(app, appname)
			return fmt.Errorf("failed to run application: %v", err)
		} else if err != nil {
			inner := err.(applicationError)
			return inner.error
		}
		return nil
	}
}

// GetFlagSet returns a flagset that corresponds to an application. This flagset can then be used
// like a *flag.FlagSet, with the additional .Stringify method.
func (commander Commander) GetFlagSet(app interface{}, appname string) (*FlagSet, error) {
	flagset := flag.NewFlagSet(appname, commander.FlagErrorHandling)
	setter := newFlagSet(flagset)
	defer setter.finish()

	err := setupFlagSet(app, setter)
	return setter, errors.Wrapf(err, "failed to get flagset")
}

// Usage returns the "help" string for this application.
func (commander Commander) Usage(app interface{}) string {
	// First use the flagset to print flags
	appname := getCLIName(app)
	return commander.NamedUsage(app, appname)
}

// NamedUsage returns the usage of the CLI application with a custom name at the top.
func (commander Commander) NamedUsage(app interface{}, appname string) string {
	var buf bytes.Buffer

	flagset, err := commander.GetFlagSet(app, appname)
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
func (commander Commander) PrintUsage(app interface{}, appname string) {
	usage := commander.NamedUsage(app, appname)
	fmt.Fprintf(commander.UsageOutput, usage)
}

func findCommand(app interface{}, commands []string) (string, error) {
	for _, cmd := range commands {
		if found, err := hasCommand(app, cmd); err != nil {
			return "", err
		} else if found {
			return cmd, nil
		}
	}
	return "", nil
}

func executeCommand(app interface{}, cmd string, args []string, flagset *flag.FlagSet) error {
	// Reparse flags to populate some of the flags that the default package might have missed
	if err := flagset.Parse(args); err != nil {
		return errors.WithStack(err)
	}
	args = flagset.Args()

	// Execute post flag parse hook
	if err := executeHook(app); err != nil {
		return errors.WithStack(err)
	}

	// Finally run that command if everything seems fine
	if err := runCommand(app, cmd, args...); err != nil {
		return err
	}
	return nil
}

// runCommand runs a specific command of the application with arguments.
func runCommand(app interface{}, cmd string, args ...string) error {
	method, err := getMethod(app, cmd)
	if err != nil {
		return err
	}

	// Make sure we have enough args for this command
	inputsize := method.Type.NumIn() - 1
	if len(args) < inputsize-1 && method.Type.In(inputsize).Kind() == reflect.Slice {
		return fmt.Errorf("command requires %v arguments, have %v", inputsize-1, len(args))
	} else if len(args) != inputsize && method.Type.In(inputsize).Kind() != reflect.Slice {
		return fmt.Errorf("command requires %v arguments, have %v", inputsize, len(args))
	} else if len(args) < inputsize {
		args = append(args, "[]")
	} else if len(args) > inputsize || method.Type.In(inputsize).Kind() == reflect.Slice {
		// Then we consider that the extra arguments are just a list of strings
		extras := args[inputsize-1:]
		bytes, _ := json.Marshal(extras)
		args[inputsize-1] = string(bytes)
		args = args[:inputsize]
	}

	in := make([]reflect.Value, inputsize+1)
	in[0] = reflect.ValueOf(app)
	for i, arg := range args {
		t := method.Type.In(i + 1)
		param, err := utils.ParseString(t, arg)
		if err != nil {
			return errors.Wrapf(err, "failed to parse string into function argument")
		}
		in[i+1] = param
	}
	out := method.Func.Call(in)
	if len(out) == 0 {
		return nil
	} else if err, ok := out[0].Interface().(error); ok {
		return applicationError{err}
	}
	return nil
}

// subCommand returns the subcommand struct that corresponds to the command cmd. If none is found,
// subCommand returns nil, nil.
func subCommand(app interface{}, cmd string) (interface{}, error) {
	st, valid := utils.DerefType(app)
	if !valid {
		return nil, fmt.Errorf("application needs to be a struct or a pointer to a struct")
	}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 && (split[0] == FlagDirective || split[0] == SubcommandDirective) {
				return nil, fmt.Errorf("malformed tag on application: %v", alias)
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
				return nil, fmt.Errorf("failed to get subcommand from field %v of type %v", field.Name, st.Name())
			}
			fieldval := v.FieldByName(field.Name)
			if !fieldval.IsValid() {
				return nil, fmt.Errorf("failed to get subcommand from field %v of type %v", field.Name, st.Name())
			}
			return fieldval.Interface(), nil
		}
	}
	return nil, nil
}

// hasCommand returns true if the application implements a specific command; and false otherwise.
func hasCommand(app interface{}, cmd string) (bool, error) {
	cmd = normalizeCommand(cmd)
	apptype := reflect.TypeOf(app)
	for i := 0; i < apptype.NumMethod(); i++ {
		method := apptype.Method(i)
		if strings.ToLower(method.Name) == cmd {
			return true, nil
		}
	}
	return false, nil
}

// setupflagSet goes through the type of the application and creates flags on the flagset passed in.
func setupFlagSet(app interface{}, setter *FlagSet) error {
	// Get the raw type of the app
	st, valid := utils.DerefType(app)
	if !valid {
		return fmt.Errorf("application needs to be a struct or a pointer to a struct")
	}

	// Look through each field for flags and subcommand flags
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
			split := strings.Split(alias, "=")
			if len(split) != 2 && (split[0] == FlagDirective || split[0] == SubcommandDirective) {
				return fmt.Errorf("malformed tag on application: %v", alias)
			}

			// If this field is itself a flag
			if split[0] == FlagDirective {
				err := setter.setFlag(app, field, split[1])
				if err != nil {
					return errors.Wrapf(err, "failed to setup flag for application")
				}
			}

			// If this field has subflags, recurse inside that
			if split[0] == FlagStructDirective && len(split) == 1 {
				v, valid := utils.DerefValue(app)
				if !valid || v.Kind() != reflect.Struct {
					// The subapp is nil or not a struct
					return nil
				}
				fieldval := v.FieldByName(field.Name)
				if !fieldval.IsValid() {
					return fmt.Errorf("failed to get flags from field %v of type %v", field.Name, st.Name())
				}
				fieldIface := fieldval.Interface()
				if fieldval.Type().Kind() == reflect.Struct {
					fieldIface = fieldval.Addr().Interface()
				}
				if err := setupFlagSet(fieldIface, setter); err != nil {
					return errors.Wrap(err, "failed to get flagset for sub-struct")
				}
			} else if split[0] == FlagSliceDirective {
				v, valid := utils.DerefValue(app)
				if !valid || v.Kind() != reflect.Struct {
					// The subapp is nil or not a struct
					return nil
				}
				fieldval := v.FieldByName(field.Name)
				if !fieldval.IsValid() {
					return fmt.Errorf("failed to get flags from field %v of type %v", field.Name, st.Name())
				} else if fieldval.Kind() != reflect.Slice {
					return fmt.Errorf("FlagSlice directive should only be used on slice fields")
				}
				for i := 0; i < fieldval.Len(); i++ {
					item := fieldval.Index(i)
					if err := setupFlagSet(item.Interface(), setter); err != nil {
						return errors.Wrap(err, "failed to get flagset for slice element")
					}
				}
			}
		}
	}
	return nil
}

// parseSubcommandDirective parses the subcommand directive into the subcommand string and its description.
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

func getCLIName(app interface{}, commands ...string) string {
	appname := "CLI"
	if casted, ok := app.(NamedCLI); ok {
		appname = casted.CLIName()
	}
	if len(commands) > 0 {
		appname += " " + strings.Join(commands, " ")
	}
	return appname
}

func normalizeCommand(cmd string) string {
	cmd = strings.Replace(cmd, "-", "", -1)
	cmd = strings.Replace(cmd, "_", "", -1)
	cmd = strings.ToLower(cmd)
	return cmd
}

func getMethod(app interface{}, cmd string) (reflect.Method, error) {
	apptype := reflect.TypeOf(app)
	var method reflect.Method
	for i := 0; i < apptype.NumMethod(); i++ {
		method = apptype.Method(i)
		if strings.ToLower(method.Name) == normalizeCommand(cmd) {
			return method, nil
		}
	}
	return method, fmt.Errorf("failed to find method %v", cmd)
}
