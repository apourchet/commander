package commander_test

import (
	"fmt"
	"testing"

	"github.com/apourchet/commander"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PetStore will have the following commands
// petstore manage init
// petstore manage copy <new-location>
// petstore manage delete
// petstore manage default <location>
// petstore add <petname>
// petstore remove <petname>
type PetStore struct {
	DryRun bool `commander:"flag=dry-run"`

	Manager PetStoreManager `commander:"subcommand=manage"`
}

func (store PetStore) Add(petname string) {}

func (store PetStore) Remove(petname string) {}

type PetStoreManager struct {
	StoreLocation string `commander:"long=store-location"`
}

func (manager PetStoreManager) Init() {}

func (manager PetStoreManager) Copy(newLocation string) {}

func (manager PetStoreManager) Delete() {}

func (manager PetStoreManager) Default(newLocation string) {}

func TestCommanderBasics(t *testing.T) {
	app := &PetStore{}
	args := []string{"-dry-run", "store-location", "/tmp/petstore"}
	err := commander.New().RunCLI(app, args)
	fmt.Println(err)
	t.Fail()
}

func TestFlagParsing(t *testing.T) {
	app := &PetStore{}
	cmd := commander.New()
	flagset, err := cmd.GetFlagSet(app)
	require.Nil(t, err)

	args := []string{"-dry-run", "store-location", "/tmp/petstore"}
	flagset.Parse(args)
	assert.True(t, app.DryRun)
}
