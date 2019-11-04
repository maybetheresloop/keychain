package main

import (
	"github.com/maybetheresloop/keychain/internal"
	"github.com/urfave/cli"
)

func Set(c *cli.Context) error {
	key := c.Args().Get(0)
	if key == "" {
		return ErrKeyNotSpecified
	}

	value := c.Args().Get(1)
	if value == "" {
		return ErrValueNotSpecified
	}

	fp := c.String("file")

	keychain, err := internal.Open(fp)
	if err != nil {
		return err
	}

	defer keychain.Close()

	if err := keychain.Set([]byte(key), []byte(value)); err != nil {
		return err
	}

	return nil
}
