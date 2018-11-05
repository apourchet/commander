package commander

import (
	"fmt"
	"reflect"

	"github.com/apourchet/commander/utils"
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
