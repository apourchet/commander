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

	app := &FlagTester{}
	flagset, err := cmd.GetFlagSet(app)
	require.NoError(t, err)
	args := []string{"--boolflag=true", "--stringflag", "somestring", "--intflag", "10"}
	flagset.Parse(args)
	require.True(t, app.Bool)
	require.Equal(t, "somestring", app.String)
	require.Equal(t, 10, app.Int)

	app = &FlagTester{}
	flagset, err = cmd.GetFlagSet(app)
	require.NoError(t, err)
	args = []string{"--boolflag", "--stringflag", "somestring", "--intflag", "10"}
	flagset.Parse(args)
	require.True(t, app.Bool)
	require.Equal(t, "somestring", app.String)
	require.Equal(t, 10, app.Int)
}

func TestFlagDefaults(t *testing.T) {
	cmd := commander.New()

	app := &FlagTester{
		String: "somestring",
		Bool:   true,
	}
	flagset, err := cmd.GetFlagSet(app)
	require.NoError(t, err)
	args := []string{"--intflag", "10"}
	flagset.Parse(args)
	require.True(t, app.Bool)
	require.Equal(t, "somestring", app.String)
	require.Equal(t, 10, app.Int)
}

type FlagTesterNested struct {
	Toplevel bool `commander:"flag=toplevel,A toplevel bool"`

	Nested *FlagTester `commander:"flagstruct"`
}

func TestFlagParsingNested(t *testing.T) {
	cmd := commander.New()

	app := &FlagTesterNested{Nested: &FlagTester{}}
	flagset, err := cmd.GetFlagSet(app)
	require.NoError(t, err)
	args := []string{"--boolflag=true", "--toplevel=true", "--stringflag", "somestring", "--intflag", "10"}
	flagset.Parse(args)
	require.Equal(t, true, app.Toplevel)
	require.Equal(t, true, app.Nested.Bool)
	require.Equal(t, 10, app.Nested.Int)
	require.Equal(t, "somestring", app.Nested.String)

	app = &FlagTesterNested{Nested: &FlagTester{}}
	flagset, err = cmd.GetFlagSet(app)
	require.NoError(t, err)
	args = []string{"--boolflag", "--toplevel", "--stringflag", "somestring", "--intflag", "10"}
	flagset.Parse(args)
	require.Equal(t, true, app.Toplevel)
	require.Equal(t, true, app.Nested.Bool)
	require.Equal(t, 10, app.Nested.Int)
	require.Equal(t, "somestring", app.Nested.String)
}

type FlagDurationTester struct {
	Duration time.Duration            `commander:"flag=dur,A duration"`
	Nested   *OtherFlagDurationTester `commander:"flagstruct"`
}

type OtherFlagDurationTester struct {
	Duration time.Duration `commander:"flag=otherdur,Another duration"`
}

func TestFlagParsingDuration(t *testing.T) {
	cmd := commander.New()

	app := &FlagDurationTester{
		Nested: &OtherFlagDurationTester{},
	}
	flagset, err := cmd.GetFlagSet(app)
	require.NoError(t, err)
	args := []string{"--dur", "4h", "--otherdur", "2s"}
	flagset.Parse(args)
	require.Equal(t, 4*time.Hour, app.Duration)
	require.Equal(t, 2*time.Second, app.Nested.Duration)
}
