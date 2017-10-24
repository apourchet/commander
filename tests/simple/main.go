package main

import (
	"fmt"
	"os"

	"github.com/apourchet/commander"
)

// PetStore will have the following commands
// petstore manage init
// petstore manage copy <new-location>
// petstore manage delete
// petstore manage default <location>
// petstore add <petname>
// petstore remove <petname>
type PetStore struct {
	DryRun bool `commander:"long=dru-run"`

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

func main() {
	err := commander.New().RunCLI(PetStore{}, os.Args)
	fmt.Println(err)
}
