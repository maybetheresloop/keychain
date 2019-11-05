package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

var DefaultDbFile = "keychain.db"

func main() {
	app := cli.NewApp()
	app.Name = "keychain"
	app.Usage = "A key/value store backed by a persistent log."
	app.Version = "0.1.0"

	fileFlag := cli.StringFlag{
		Name:      "file, f",
		Value:     DefaultDbFile,
		Usage:     "Database `FILE` to use",
		TakesFile: true,
	}

	app.Commands = []cli.Command{
		{
			Name:   "get",
			Usage:  "retrieve a value from the store",
			Action: Get,
			Flags: []cli.Flag{
				fileFlag,
			},
		},
		{
			Name:   "set",
			Usage:  "set a key/value pair in the store",
			Action: Set,
			Flags: []cli.Flag{
				fileFlag,
			},
		},
		{
			Name:   "rm",
			Usage:  "remove a key/value pair from the store",
			Action: Remove,
			Flags: []cli.Flag{
				fileFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
