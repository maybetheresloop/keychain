package main

import (
	"fmt"
	"os"

	"github.com/maybetheresloop/keychain/internal/util"

	"github.com/maybetheresloop/keychain/internal"
	"github.com/urfave/cli"
)

func Remove(c *cli.Context) error {
	key := c.Args().Get(0)
	if key == "" {
		return ErrKeyNotSpecified
	}

	fp := c.String("file")

	if _, err := os.Stat(fp); err != nil {
		return err
	}

	keychain, err := internal.Open(fp)
	if err != nil {
		return err
	}

	defer keychain.Close()

	removed, err := keychain.Remove([]byte(key))
	if err != nil {
		return err
	}

	if removed {
		fmt.Printf("%s %d\n", util.Integer, 1)
	} else {
		fmt.Printf("%s %d\n", util.Integer, 0)
	}

	return nil
}
