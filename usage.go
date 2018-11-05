package commander

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/apourchet/commander/utils"
)

// Usage returns the "help" string for this application.
func (commander Commander) Usage(app interface{}) string {
	appname := getCLIName(app)
	return commander.NamedUsage(app, appname)
}

// UsageWithCommand returns the usage of this application given the command passed in.
func (commander Commander) UsageWithCommand(app interface{}, cmd string) string {
	appname := getCLIName(app)
	return commander.NamedUsageWithCommand(app, appname, cmd)
}

// NamedUsage returns the usage of the CLI application with a custom name at the top.
func (commander Commander) NamedUsage(app interface{}, appname string) string {
	flagset, _ := commander.GetFlagSet(app, appname)
	return usageWithFlagset(app, flagset)
}

// NamedUsageWithCommand returns the usage of this application given the command passed in, with
// a custom name at the top.
func (commander Commander) NamedUsageWithCommand(app interface{}, appname string, cmd string) string {
	flagset, _ := commander.GetFlagSetWithCommand(app, appname, cmd)
	return usageWithFlagset(app, flagset)
}

// PrintUsage prints the usage of the application given to the io.Writer specified; unless the
// Commander fails to get the usage for this application.
func (commander Commander) PrintUsage(app interface{}, appname string) {
	usage := commander.NamedUsage(app, appname)
	fmt.Fprintf(commander.UsageOutput, usage)
}

// PrintUsageWithCommand prints the usage of the application like PrintUsage but for the specific
// subcommand provided.
func (commander Commander) PrintUsageWithCommand(app interface{}, appname string, cmd string) {
	usage := commander.NamedUsageWithCommand(app, appname, cmd)
	fmt.Fprintf(commander.UsageOutput, usage)
}

func usageWithFlagset(app interface{}, flagset *FlagSet) string {
	var buf bytes.Buffer
	if flagset != nil {
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
