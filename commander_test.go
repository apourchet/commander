package commander_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/apourchet/commander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Application struct {
	count int

	IntFlag int `commander:"flag=intflag,An int"`

	SubApp  *SubApplication `commander:"subcommand=subapp,Use subapp commands"`
	SubApp2 *SubApplication `commander:"subcommand=subapp2,Use subapp commands"`
}

func (app *Application) OpOne(str string) {
	if str == "test" {
		app.count++
	}
}

func (app *Application) OpTwo(i int) {
	if i == 30 {
		app.count++
	}
}

func (app *Application) CLIName() string { return "myapp" }

func (app *Application) OpVariadic(name string, names []string) {
	app.count += len(names)
}

type SubApplication struct {
	count int

	SubIntFlag int `commander:"flag=subintflag,Another int"`

	SubSubApp *SubSubApplication `commander:"subcommand=subsubapp,Use subsubapp commands"`
}

func (app *SubApplication) OpThree() {
	app.count++
}

func (app *SubApplication) OpFour(m map[string]string) {
	if m["test"] == "testing" {
		app.count++
	}
}

type SubSubApplication struct {
	count int
}

func (app *SubSubApplication) OpDeep() {
	app.count++
}

func TestCommanderBasics(t *testing.T) {
	cmd := commander.New()
	cmd.UsageOutput = ioutil.Discard

	app := &Application{}
	args := []string{"opone", "test"}
	err := cmd.RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.count)

	args = []string{"opone"}
	err = cmd.RunCLI(app, args)
	require.Error(t, err)
}

func TestCommanderInt(t *testing.T) {
	app := &Application{}
	args := []string{"optwo", "30"}
	err := commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.count)
}

func TestCommanderVariadic(t *testing.T) {
	app := &Application{count: -5}
	args := []string{"opvariadic", "a"}
	err := commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, -5, app.count)

	args = []string{"opvariadic", "a", "b"}
	err = commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, -4, app.count)

	args = []string{"opvariadic", "a", "b", "c", "d"}
	err = commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, -1, app.count)
}

func TestCommanderSubcommand(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{},
	}
	args := []string{"subapp", "opthree"}
	err := commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.count)
}

func TestSubcommandArguments(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{},
	}
	args := []string{"subapp", "opfour", `{"test": "testing"}`}
	err := commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.count)
}

func TestSubSubcommand(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{
			SubSubApp: &SubSubApplication{},
		},
	}
	args := []string{"subapp", "subsubapp", "opdeep"}
	err := commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.SubSubApp.count)
}

func TestFlagOrder(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{},
	}
	args := []string{"--intflag", "11", "opone", "--intflag", "10", "test"}
	err := commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.count)
	require.Equal(t, 10, app.IntFlag)

	app = &Application{
		SubApp: &SubApplication{},
	}
	args = []string{"--intflag", "10", "subapp", "opthree"}
	err = commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.count)
	require.Equal(t, 10, app.IntFlag)
	require.Equal(t, 0, app.SubApp.SubIntFlag)

	app = &Application{
		SubApp: &SubApplication{},
	}
	args = []string{"subapp", "--subintflag", "10", "opthree"}
	err = commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.count)
	require.Equal(t, 0, app.IntFlag)
	require.Equal(t, 10, app.SubApp.SubIntFlag)

	app = &Application{
		SubApp: &SubApplication{},
	}
	args = []string{"--intflag", "10", "subapp", "--subintflag", "10", "opthree"}
	err = commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.count)
	require.Equal(t, 10, app.IntFlag)
	require.Equal(t, 10, app.SubApp.SubIntFlag)
}

func TestUsage(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{},
	}
	cmd := commander.New()
	expected := `Usage of myapp:
  -intflag value
    	An int (default=0)

Sub-Commands:
  subapp  |  Use subapp commands
  subapp2  |  Use subapp commands
`
	usage := cmd.Usage(app)
	AssertEqualLines(t, expected, usage)
}

func AssertEqualLines(t *testing.T, expected, actual string) {
	swapped := false
	small, big := strings.Split(expected, "\n"), strings.Split(actual, "\n")
	if len(small) > len(big) {
		small, big = big, small
		swapped = true
	}

	for i := range small {
		if !swapped {
			assert.Equal(t, small[i], big[i])
		} else {
			assert.Equal(t, big[i], small[i])
		}
	}

	symbol := "+"
	if swapped {
		symbol = "-"
	}

	for i := len(small); i < len(big); i++ {
		assert.Fail(t, symbol+big[i])
	}
}
