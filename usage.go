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

	directives := map[string]string{}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup(FieldTag); ok && alias != "" {
			split := strings.SplitN(alias, "=", 2)
			if len(split) != 2 {
				continue
			} else if split[0] != FlagStructDirective &&
				split[0] != SubcommandDirective {
				continue
			}

			cmd, newdesc := parseSubcommandDirective(split[1])
			if split[0] == FlagStructDirective {
				if found, _ := hasCommand(app, cmd); !found {
					continue
				}
			}

			if desc, found := directives[cmd]; !found || desc == "" {
				directives[cmd] = newdesc
			}
		}
	}

	if len(directives) == 0 {
		return buf.String()
	}

	fmt.Fprintf(&buf, "\nSub-Commands:\n")
	cmds := sortKeys(directives)
	for _, cmd := range cmds {
		desc := "No description for this subcommand"
		if directives[cmd] != "" {
			desc = directives[cmd]
		}
		if provider, ok := app.(CommandDescriptionProvider); ok {
			if newdesc := provider.GetCommandDescription(cmd); newdesc != "" {
				desc = newdesc
			}
		}
		fmt.Fprintf(&buf, "  %v  |  %v\n", cmd, desc)
	}

	return buf.String()
}
