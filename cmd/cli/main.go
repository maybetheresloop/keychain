package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maybetheresloop/keychain/pkg/resp"
	"github.com/urfave/cli"
)

func prompt(sc *bufio.Scanner, addr string) bool {
	fmt.Printf("%s> ", addr)
	return sc.Scan()
}

func runCmd(cmd string, tail []string) error {
	return nil
}

func run(c *cli.Context) error {
	if c.NArg() > 0 {
		return runCmd(c.Args().First(), c.Args().Tail())
	}

	return runCli(c)
}

func runCli(c *cli.Context) error {
	fmt.Println(c.NArg())

	fmt.Println("Welcome to keychain-cli!")

	addr := fmt.Sprintf("%s:%d", c.String("host"), c.Uint("port"))

	//_, err := net.Dial("tcp", addr)
	//if err != nil {
	//	return err
	//}
	w := resp.NewWriter(os.Stdout)

	sc := bufio.NewScanner(os.Stdin)
	for prompt(sc, addr) {
		line := sc.Text()

		message := strings.Split(line, " ")
		w.WriteMessage(message)
		w.Flush()
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
