package commander_test

import (
	"testing"
	"time"

	"github.com/apourchet/commander"
	"github.com/stretchr/testify/require"
)

type FlagTester struct {
	String string `commander:"flag=stringflag,A string"`
	Int    int    `commander:"flag=intflag,An int"`
	Bool   bool   `commander:"flag=boolflag,A bool"`
}

func TestFlagParsing(t *testing.T) {
	cmd := commander.New()

	t.Run("1", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{"--boolflag=true", "--stringflag", "somestring", "--intflag", "10"}
		flagset.Parse(args)
		require.True(t, app.Bool)
		require.Equal(t, "somestring", app.String)
		require.Equal(t, 10, app.Int)
	})

	t.Run("2", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{"--boolflag", "--stringflag", "somestring", "--intflag", "10"}
		flagset.Parse(args)
		require.True(t, app.Bool)
		require.Equal(t, "somestring", app.String)
		require.Equal(t, 10, app.Int)
	})

	t.Run("3", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{"--stringflag", "somestring", "--intflag", "10"}
		flagset.Parse(args)
		require.False(t, app.Bool)
		require.Equal(t, "somestring", app.String)
		require.Equal(t, 10, app.Int)
	})
}

func TestFlagStringify(t *testing.T) {
	cmd := commander.New()

	t.Run("1", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{"--boolflag", "--stringflag", "somestring", "--intflag", "10"}
		flagset.Parse(args)
		newargs := flagset.Stringify()

		app = &FlagTester{}
		flagset, err = cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		flagset.Parse(newargs)
		require.True(t, app.Bool)
		require.Equal(t, "somestring", app.String)
		require.Equal(t, 10, app.Int)
	})

	t.Run("2", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{"--boolflag", "--stringflag", "somestring", "--intflag", "10"}
		flagset.Parse(args)
		newargs := flagset.Stringify()

		app = &FlagTester{}
		flagset, err = cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		flagset.Parse(newargs)
		require.True(t, app.Bool)
		require.Equal(t, "somestring", app.String)
		require.Equal(t, 10, app.Int)
	})

	t.Run("3", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{"--stringflag", "somestring", "--intflag", "10"}
		flagset.Parse(args)
		newargs := flagset.Stringify()

		app = &FlagTester{}
		flagset, err = cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		flagset.Parse(newargs)
		require.False(t, app.Bool)
		require.Equal(t, "somestring", app.String)
		require.Equal(t, 10, app.Int)
	})

	t.Run("4", func(t *testing.T) {
		app := &FlagTester{}
		flagset, err := cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		args := []string{}
		flagset.Parse(args)
		newargs := flagset.Stringify()

		app = &FlagTester{}
		flagset, err = cmd.GetFlagSet(app, "CLI")
		require.NoError(t, err)
		flagset.Parse(newargs)
		require.False(t, app.Bool)
		require.Equal(t, "", app.String)
		require.Equal(t, 0, app.Int)
	})
}

func TestFlagDefaults(t *testing.T) {
	cmd := commander.New()

	app := &FlagTester{
		String: "somestring",
		Bool:   true,
	}
	flagset, err := cmd.GetFlagSet(app, "CLI")
	require.NoError(t, err)
	args := []string{"--intflag", "10"}
	flagset.Parse(args)
	require.True(t, app.Bool)
	require.Equal(t, "somestring", app.String)
	require.Equal(t, 10, app.Int)
}

type FlagTesterNested struct {
	Toplevel bool `commander:"flag=toplevel,A toplevel bool"`

	Nested      *FlagTester `commander:"flagstruct"`
	NestedNoPtr struct {
		Int int `commander:"flag=innerint"`
	} `commander:"flagstruct"`
}

func TestFlagParsingNested(t *testing.T) {
	cmd := commander.New()

	app := &FlagTesterNested{Nested: &FlagTester{}}
	flagset, err := cmd.GetFlagSet(app, "CLI")
	require.NoError(t, err)
	args := []string{"--boolflag=true", "--toplevel=true", "--stringflag", "somestring", "--intflag", "10"}
	flagset.Parse(args)
	require.Equal(t, true, app.Toplevel)
	require.Equal(t, true, app.Nested.Bool)
	require.Equal(t, 10, app.Nested.Int)
	require.Equal(t, "somestring", app.Nested.String)

	app = &FlagTesterNested{Nested: &FlagTester{}}
	flagset, err = cmd.GetFlagSet(app, "CLI")
	require.NoError(t, err)
	args = []string{"--boolflag", "--toplevel", "--stringflag", "somestring", "--intflag", "10", "--innerint=10"}
	flagset.Parse(args)
	require.Equal(t, true, app.Toplevel)
	require.Equal(t, true, app.Nested.Bool)
	require.Equal(t, 10, app.Nested.Int)
	require.Equal(t, "somestring", app.Nested.String)
	require.Equal(t, 10, app.NestedNoPtr.Int)
}

type FlagDurationTester struct {
	Duration time.Duration           `commander:"flag=dur,A duration"`
	Nested   OtherFlagDurationTester `commander:"flagstruct"`
}

type OtherFlagDurationTester struct {
	Duration time.Duration `commander:"flag=otherdur,Another duration"`
}

func TestFlagParsingDuration(t *testing.T) {
	cmd := commander.New()

	app := &FlagDurationTester{
		Nested: OtherFlagDurationTester{},
	}
	flagset, err := cmd.GetFlagSet(app, "CLI")
	require.NoError(t, err)
	args := []string{"--dur", "4h", "--otherdur", "2s"}
	flagset.Parse(args)
	require.Equal(t, 4*time.Hour, app.Duration)
	require.Equal(t, 2*time.Second, app.Nested.Duration)
}

type FlagTesterSliced struct {
	Slice []interface{} `commander:"flagslice"`
}

type IntFlagStruct struct {
	Value int `commander:"flag=intflag2,An int"`
}

type BoolFlagStruct struct {
	Value bool `commander:"flag=boolflag2,A bool"`
}

func TestFlagParsingSliced(t *testing.T) {
	cmd := commander.New()

	intflag := &IntFlagStruct{}
	boolflag := &BoolFlagStruct{}
	app := &FlagTesterSliced{
		Slice: []interface{}{intflag, boolflag},
	}
	flagset, err := cmd.GetFlagSet(app, "CLI")
	require.NoError(t, err)
	args := []string{"--intflag2", "10", "--boolflag2"}
	flagset.Parse(args)
	require.Equal(t, 10, intflag.Value)
	require.True(t, boolflag.Value)
}
