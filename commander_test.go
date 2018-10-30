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
	count          int
	postFlagHooked bool

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

func (app *Application) PostFlagParse() error {
	app.postFlagHooked = (app.IntFlag == 10)
	return nil
}

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

type Application2 struct {
	SubCmd *SubCmd2 `commander:"subcommand=subcmd2"`
}

type SubCmd2 struct {
	Fl int `commander:"flag=anint"`

	count int
}

func (sub *SubCmd2) CommanderDefault(arg string) {
	if arg == "arg" {
		sub.count++
	}
}

func (sub *SubCmd2) Cmd1(first string, others []string) {
	if first == "first" && len(others) == 2 {
		sub.count++
	}
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
	require.True(t, app.postFlagHooked)

	app = &Application{
		SubApp: &SubApplication{},
	}
	args = []string{"--intflag", "10", "subapp", "opthree"}
	err = commander.New().RunCLI(app, args)
	require.NoError(t, err)
	require.Equal(t, 1, app.SubApp.count)
	require.Equal(t, 10, app.IntFlag)
	require.Equal(t, 0, app.SubApp.SubIntFlag)
	require.True(t, app.postFlagHooked)

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
	t.Run("top_level", func(t *testing.T) {
		app := &Application{
			IntFlag: 10,
			SubApp:  &SubApplication{},
		}
		cmd := commander.New()
		expected := `Usage of myapp:
  -intflag
    	An int (type: int, default: 10)

Sub-Commands:
  subapp  |  Use subapp commands
  subapp2  |  Use subapp commands
`
		usage := cmd.Usage(app)
		assertEqualLines(t, expected, usage)
	})
	t.Run("no_subcommand", func(t *testing.T) {
		cmd := commander.New()
		expected := `Usage of CLI:
  -anint
    	No usage found for this flag. (type: int, default: 0)
`
		usage := cmd.Usage(&SubCmd2{})
		assertEqualLines(t, expected, usage)
	})
	t.Run("empty_string_default", func(t *testing.T) {
		app := &struct {
			Str string `commander:"flag=str"`
		}{}
		expected := `Usage of CLI:
  -str
    	No usage found for this flag. (type: string, default: "")
`
		usage := commander.New().Usage(app)
		assertEqualLines(t, expected, usage)
	})
	t.Run("default_strings_types", func(t *testing.T) {
		app := &struct {
			Bool bool              `commander:"flag=b,A bool"`
			Str  string            `commander:"flag=str"`
			Strs []string          `commander:"flag=strs"`
			Map  map[string]string `commander:"flag=map"`
		}{
			Bool: true,
			Map:  map[string]string{"a": "b"},
		}
		expected := `Usage of CLI:
  -b	A bool (type: bool, default: true)
  -map
    	No usage found for this flag. (type: map, default: {"a":"b"})
  -str
    	No usage found for this flag. (type: string, default: "")
  -strs
    	No usage found for this flag. (type: slice, default: null)
`
		usage := commander.New().Usage(app)
		assertEqualLines(t, expected, usage)
	})
}

func assertEqualLines(t *testing.T, expected, actual string) {
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

func TestApplication2(t *testing.T) {
	t.Run("calls_commander_default", func(t *testing.T) {
		app := &Application2{
			SubCmd: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "arg"})
		require.NoError(t, err)
		require.Equal(t, 1, app.SubCmd.count)
	})
	t.Run("not_enough_arguments", func(t *testing.T) {
		app := &Application2{
			SubCmd: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2"})
		require.Error(t, err)
	})
	t.Run("too_many_arguments", func(t *testing.T) {
		app := &Application2{
			SubCmd: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "arg", "arg2"})
		require.Error(t, err)
	})
	t.Run("last_arg_string_array", func(t *testing.T) {
		app := &Application2{
			SubCmd: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "cmd1", "first", "second", "third"})
		require.NoError(t, err)
		require.Equal(t, 1, app.SubCmd.count)
	})
}
