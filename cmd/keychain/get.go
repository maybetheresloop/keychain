package main

import (
	"fmt"
	"os"

	"github.com/maybetheresloop/keychain"

	"github.com/maybetheresloop/keychain/internal/util"
	"github.com/urfave/cli"
)

func Get(c *cli.Context) error {
	key := c.Args().Get(0)
	if key == "" {
		return ErrKeyNotSpecified
	}

	fp := c.String("file")

	if _, err := os.Stat(fp); err != nil {
		return err
	}

	keys, err := keychain.Open(fp)
	if err != nil {
		return err
	}

	defer keys.Close()

	value, err := keys.Get([]byte(key))
	if err != nil {
		return err
	}

	if value == nil {
		fmt.Println(util.Nil)
	} else {
		fmt.Printf("%q\n", string(value))
	}

	return nil
}
