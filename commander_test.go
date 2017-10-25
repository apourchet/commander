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

func (app Application) OpOne(str string) {
	if str == "test" {
		app.count++
	}
}

func (app Application) OpTwo(i int) {}

type SubApplication struct {
	count int
}

func (app SubApplication) OpThree() {}

func (app SubApplication) OpFour(m map[string]string) {
	if m["test"] == "testing" {
		app.count++
	}
}

func TestCommanderBasics(t *testing.T) {
	app := &Application{}
	args := []string{"opone"}
	err := commander.New().RunCLI(app, args)
	assert.Nil(t, err)
	assert.Equal(t, 1, app.count)
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
