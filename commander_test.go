package commander_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/apourchet/commander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommanderBasics(t *testing.T) {
	cmd := commander.New()
	cmd.UsageOutput = ioutil.Discard
	app := &Application{}

	t.Run("1", func(t *testing.T) {
		args := []string{"opone", "test"}
		err := cmd.RunCLI(app, args)
		require.NoError(t, err)
		require.Equal(t, 1, app.count)
	})

	t.Run("2", func(t *testing.T) {
		args := []string{"opone"}
		err := cmd.RunCLI(app, args)
		require.Error(t, err)
	})

	t.Run("3", func(t *testing.T) {
		args := []string{"opthree"}
		err := cmd.RunCLI(app, args)
		require.Error(t, err)
		require.Equal(t, "ERROR", err.Error())
		require.Equal(t, errTest, err)
	})
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
	t.Run("1", func(t *testing.T) {
		app := &Application{
			SubApp: &SubApplication{},
		}
		args := []string{"subapp", "opthree"}
		err := commander.New().RunCLI(app, args)
		require.NoError(t, err)
		require.Equal(t, 1, app.SubApp.count)
	})

	t.Run("2", func(t *testing.T) {
		app := &Application{
			SubApp: &SubApplication{},
		}
		args := []string{"subapp", "test"}
		err := commander.New().RunCLI(app, args)
		require.Equal(t, errTest, err)
	})
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
	t.Run("1", func(t *testing.T) {
		app := &Application{
			SubApp: &SubApplication{},
		}
		args := []string{"--intflag", "10", "subapp", "--subintflag", "10", "test"}
		err := commander.New().RunCLI(app, args)
		require.Error(t, err)
		require.Equal(t, errTest, err)
		require.True(t, app.postFlagHooked)
	})

	t.Run("2", func(t *testing.T) {
		app := &Application{
			SubApp: &SubApplication{},
		}
		args := []string{"--intflag", "10", "subapp", "opthree"}
		err := commander.New().RunCLI(app, args)
		require.NoError(t, err)
		require.Equal(t, 1, app.SubApp.count)
		require.Equal(t, 10, app.IntFlag)
		require.Equal(t, 0, app.SubApp.SubIntFlag)
		require.True(t, app.postFlagHooked)
	})

	t.Run("3", func(t *testing.T) {
		app := &Application{
			SubApp: &SubApplication{},
		}
		args := []string{"subapp", "--subintflag", "10", "opthree"}
		err := commander.New().RunCLI(app, args)
		require.NoError(t, err)
		require.Equal(t, 1, app.SubApp.count)
		require.Equal(t, 0, app.IntFlag)
		require.Equal(t, 10, app.SubApp.SubIntFlag)

	})

	t.Run("4", func(t *testing.T) {
		app := &Application{
			SubApp: &SubApplication{},
		}
		args := []string{"--intflag", "10", "subapp", "--subintflag", "10", "opthree"}
		err := commander.New().RunCLI(app, args)
		require.NoError(t, err)
		require.Equal(t, 1, app.SubApp.count)
		require.Equal(t, 10, app.IntFlag)
		require.Equal(t, 10, app.SubApp.SubIntFlag)
	})
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
    	An int, with a comma in the description and an = in there too (type: int, default: 10)

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

func TestApplication2(t *testing.T) {
	t.Run("calls_commander_default", func(t *testing.T) {
		app := &Application2{
			SubCmd2: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "arg"})
		require.NoError(t, err)
	})
	t.Run("calls_subcmd_name", func(t *testing.T) {
		app := &Application2{
			SubCmd: &SubCmd{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd", "subcmd", "subcmd"})
		require.NoError(t, err)
	})
	t.Run("not_enough_arguments", func(t *testing.T) {
		app := &Application2{
			SubCmd2: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2"})
		require.Error(t, err)
	})
	t.Run("too_many_arguments", func(t *testing.T) {
		app := &Application2{
			SubCmd2: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "arg", "arg2"})
		require.Error(t, err)
	})
	t.Run("last_arg_string_array", func(t *testing.T) {
		app := &Application2{
			SubCmd2: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "cmd1", "first", "second", "third"})
		require.NoError(t, err)
	})
	t.Run("array_not_enough_args", func(t *testing.T) {
		app := &Application2{
			SubCmd2: &SubCmd2{},
		}
		err := commander.New().RunCLI(app, []string{"subcmd2", "cmd1"})
		require.Error(t, err)
		require.NotEqual(t, errTest, err)

		err = commander.New().RunCLI(app, []string{"subcmd2", "cmd1", "first"})
		require.Equal(t, errTest, err)
	})
}

func TestApplication3(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		app := &Application3{}
		args := []string{"--common", "1"}
		err := commander.New().RunCLI(app, args)
		require.Error(t, err)
	})

	t.Run("2", func(t *testing.T) {
		app := &Application3{}
		args := []string{"cmd1", "--common", "1", "--b2", "1", "arg1"}
		err := commander.New().RunCLI(app, args)
		require.NoError(t, err)
	})

	t.Run("3", func(t *testing.T) {
		app := &Application3{}
		args := []string{"cmd1", "--c2", "1"}
		err := commander.New().RunCLI(app, args)
		require.Error(t, err)
	})

	t.Run("usage", func(t *testing.T) {
		expected := `Usage of CLI cmd1:
  -b2
    	No usage found for this flag. (type: string, default: "")
  -common
    	No usage found for this flag. (type: string, default: "")

Sub-Commands:
  cmd1  |  Runs cmd1
  cmd2  |  No description for this subcommand
`
		buf := &bytes.Buffer{}
		cmd := commander.New()
		cmd.UsageOutput = buf
		err := cmd.RunCLI(&Application3{}, []string{"cmd1"})
		require.Error(t, err)
		assertEqualLines(t, expected, buf.String())
	})

	t.Run("usage_2", func(t *testing.T) {
		expected := `flag provided but not defined: -asd
Usage of CLI cmd1:
  -b2
    	No usage found for this flag. (type: string, default: "")
  -common
    	No usage found for this flag. (type: string, default: "")
`
		buf := &bytes.Buffer{}
		cmd := commander.New()
		cmd.UsageOutput = buf
		err := cmd.RunCLI(&Application3{}, []string{"cmd1", "--asd"})
		require.Error(t, err)
		assertEqualLines(t, expected, buf.String())
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
