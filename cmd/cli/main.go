package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maybetheresloop/keychain"

	"github.com/urfave/cli"
)

func prompt(sc *bufio.Scanner, addr string) bool {
	fmt.Printf("%s> ", addr)
	return sc.Scan()
}

func runCmd(c client, cmd string, tail []string) error {
	switch strings.ToLower(cmd) {
	case "set":
		res, err := c.Set(tail[0], tail[1])
		if err != nil {
			return err
		}

		switch v := res.(type) {
		case string:
			fmt.Printf("%s\n", v)
		case []byte:
			fmt.Println("(nil)")
		default:
			panic("unexpected")
		}
	case "get":
		res, err := c.Get(tail[0])
		if err != nil {
			return err
		}

		if res == nil {
			fmt.Println("(nil)")
		} else {
			fmt.Printf("%s\n", res)
		}
	case "del":
		res, err := c.Del(tail...)
		if err != nil {
			return err
		}

		fmt.Println(res)
	}

	return nil
}

func run(ctx *cli.Context) error {
	var (
		c    client
		name string
	)

	if name = "keychain.db"; name != "" {
		keys, err := keychain.Open(name)
		if err != nil {
			return err
		}
		defer keys.Close()

		c = &localClient{
			keys:   keys,
			dbName: name,
		}
	} else {
		name = fmt.Sprintf("%s:%d", ctx.String("host"), ctx.Uint("port"))
	}

	if ctx.NArg() > 0 {
		return runCmd(c, ctx.Args().First(), ctx.Args().Tail())
	}

	return runCli(c, name)
}

func runCli(c client, name string) error {
	fmt.Println("Welcome to keychain-cli!")

	sc := bufio.NewScanner(os.Stdin)
	prompt := func(sc *bufio.Scanner) bool {
		fmt.Printf("%s> ", name)
		return sc.Scan()
	}

	for prompt(sc) {
		line := sc.Text()

		parts := strings.Split(line, " ")
		if err := runCmd(c, parts[0], parts[1:]); err != nil {
			return err
		}
	}

	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "keychain-cli"
	app.Email = "maybetheresloop@gmail.com"
	app.Author = "maybetheresloop"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Usage: "The ADDRESS of the host to connect to",
			Value: "127.0.0.1",
		},
		cli.UintFlag{
			Name:  "port, p",
			Usage: "The PORT to connect to",
			Value: 7878,
		},
	}

	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
