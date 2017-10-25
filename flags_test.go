package commander_test

import (
	"testing"

	"github.com/apourchet/commander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FlagTester struct {
	String string `commander:"flag=stringflag,A string"`
	Int    int    `commander:"flag=intflag,An int"`
	Bool   bool   `commander:"flag=boolflag,A bool"`
}

func TestFlagParsing(t *testing.T) {
	app := &FlagTester{}
	cmd := commander.New()
	flagset, err := cmd.GetFlagSet(app)
	require.Nil(t, err)

	args := []string{"-boolflag=true", "-stringflag", "somestring", "-intflag", "10"}
	flagset.Parse(args)
	assert.True(t, app.Bool)
	assert.Equal(t, "somestring", app.String)
	assert.Equal(t, 10, app.Int)
}

type FlagTesterNested struct {
	Toplevel bool `commander:"flag=toplevel,A toplevel bool"`

	Nested *FlagTester `commander:"subcommand=nested"`
}

func TestFlagParsingNested(t *testing.T) {
	app := &FlagTesterNested{
		Nested: &FlagTester{},
	}
	cmd := commander.New()
	flagset, err := cmd.GetFlagSet(app)
	require.Nil(t, err)

	args := []string{"-boolflag", "true", "-toplevel", "true", "-stringflag", "somestring", "-intflag", "10"}
	flagset.Parse(args)
	assert.Equal(t, true, app.Toplevel)
	assert.Equal(t, true, app.Nested.Bool)
	assert.Equal(t, 10, app.Nested.Int)
}
