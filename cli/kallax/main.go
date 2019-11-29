package main

import (
	"fmt"
	"os"

	"github.com/zbyte/go-kallax/generator/cli/kallax/cmd"

	"github.com/urfave/cli"
)

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "kallax"
	app.Version = "1.3.9"
	app.Usage = "generate kallax models"
	app.Flags = cmd.Generate.Flags
	app.Action = cmd.Generate.Action
	app.Commands = cli.Commands{
		cmd.Generate,
		cmd.Migrate,
	}

	return app
}
