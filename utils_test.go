package commander_test

import "fmt"

type Application struct {
	count          int
	postFlagHooked bool

	IntFlag int `commander:"flag=intflag,An int"`

	SubApp  *SubApplication `commander:"subcommand=subapp,Use subapp commands"`
	SubApp2 *SubApplication `commander:"subcommand=subapp2,Use subapp commands"`
}

var errTest = fmt.Errorf("ERROR")

func (app *Application) OpOne(str string) error {
	if str == "test" {
		app.count++
	}
	return nil
}

func (app *Application) OpTwo(i int) {
	if i == 30 {
		app.count++
	}
}

func (app *Application) OpThree() error {
	return errTest
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

func (app *SubApplication) SubApp(arg string) error {
	return errTest
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
	SubCmd  *SubCmd  `commander:"subcommand=subcmd"`
	SubCmd2 *SubCmd2 `commander:"subcommand=subcmd2"`
}

type SubCmd struct {
}

func (sub *SubCmd) SubCmd(str1, str2 string) error {
	if str1 != str2 {
		return errTest
	}
	return nil
}

type SubCmd2 struct {
	Fl int `commander:"flag=anint"`
}

func (sub *SubCmd2) CommanderDefault(arg string) error {
	if arg != "arg" {
		return errTest
	}
	return nil
}

func (sub *SubCmd2) Cmd1(first string, others []string) error {
	if first != "first" || len(others) != 2 {
		return errTest
	}
	return nil
}

type Application3 struct {
	A string `commander:"flag=a"`
	B struct {
		B1 string `commander:"flag=common"`
		B2 string `commander:"flag=b2"`
	} `commander:"flagstruct=cmd1"`
	C struct {
		C1 string `commander:"flag=common"`
		C2 string `commander:"flag=c2"`
	} `commander:"flagstruct=cmd2"`
}

func (app *Application3) Cmd1(a string) error { return nil }

func (app *Application3) Cmd2(b int) error { return nil }
