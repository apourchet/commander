package commander_test

import (
	"testing"

	"github.com/apourchet/commander"
	"github.com/stretchr/testify/assert"
)

// PetStore will have the following commands
// petstore manage init
// petstore manage copy <new-location>
// petstore manage delete
// petstore manage default <location>
// petstore add <petname>
// petstore remove <petname>
type Application struct {
	SubApp *SubApplication `commander:"subcommand=subapp,Use subapp commands"`

	count int
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

func (app *Application) OpVariadic(name string, names []string) {
	if len(names) == 3 {
		app.count++
	} else {
		app.count--
	}
}

type SubApplication struct {
	count int

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
	app := &Application{}
	args := []string{"opone", "test"}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, 1, app.count)
}

func TestCommanderInt(t *testing.T) {
	app := &Application{}
	args := []string{"optwo", "30"}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, 1, app.count)
}

func TestCommanderVariadic(t *testing.T) {
	app := &Application{}
	args := []string{"opvariadic", "a"}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, -1, app.count)

	args = []string{"opvariadic", "a", "b"}
	err = commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, -2, app.count)

	args = []string{"opvariadic", "a", "b", "c", "d"}
	err = commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, -1, app.count)
}

func TestCommanderSubcommand(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{},
	}
	args := []string{"subapp", "opthree"}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, 1, app.SubApp.count)
}

func TestSubcommandArguments(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{},
	}
	args := []string{"subapp", "opfour", `{"test": "testing"}`}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, 1, app.SubApp.count)
}

func TestSubSubcommand(t *testing.T) {
	app := &Application{
		SubApp: &SubApplication{
			SubSubApp: &SubSubApplication{},
		},
	}
	args := []string{"subapp", "subsubapp", "opdeep"}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, 1, app.SubApp.SubSubApp.count)
}
