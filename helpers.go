package commander

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apourchet/commander/utils"
	"github.com/pkg/errors"
)

func getPossibleCommands(arguments, cumulativeCommands []string) []string {
	commands := []string{}
	if len(cumulativeCommands) > 0 {
		prevCmd := cumulativeCommands[len(cumulativeCommands)-1]
		commands = []string{prevCmd}
	}
	if len(arguments) > 0 {
		commands = append([]string{arguments[0]}, commands...)
	}
	return append(commands, DefaultCommand)
}

func derefFlagStruct(app interface{}, st reflect.Type, field reflect.StructField) (interface{}, error) {
	v, valid := utils.DerefValue(app)
	if !valid || v.Kind() != reflect.Struct {
		// The subapp is nil or not a struct
		return nil, nil
	}
	fieldval := v.FieldByName(field.Name)
	if !fieldval.IsValid() {
		return nil, fmt.Errorf("failed to get flags from field %v of type %v", field.Name, st.Name())
	}
	fieldIface := fieldval.Interface()
	if fieldval.Type().Kind() == reflect.Struct {
		fieldIface = fieldval.Addr().Interface()
	}
	return fieldIface, nil
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
