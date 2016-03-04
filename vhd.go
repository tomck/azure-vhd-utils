package main

import (
	"github.com/codegangsta/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "vhd"
	app.Usage = "Commands to manage VHDs"


	// global level flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show more output",
		},
	}

	app.Commands = []cli.Command{
		vhdInspectCmdHandler(),
		vhdUploadCmdHandler(),
		vhdTestCmdHandler(),
	}

	app.Run(os.Args)
}