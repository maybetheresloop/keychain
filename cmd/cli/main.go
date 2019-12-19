package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/maybetheresloop/keychain/pkg/resp"

	"github.com/maybetheresloop/keychain"

	"github.com/urfave/cli"
)

type handler func(*client, []string) (interface{}, error)

type command struct {
	argc    int
	handler handler
}

// Command map for local CLI usage only.
var commands = map[string]*command{
	"get": {
		argc:    1,
		handler: get,
	},
	"set": {
		argc:    2,
		handler: set,
	},
	"del": {
		argc:    -1,
		handler: del,
	},
	"merge": {
		argc:    0,
		handler: merge,
	},
}

func prompt(sc *bufio.Scanner, format string, args ...interface{}) bool {
	fmt.Printf(format, args...)
	return sc.Scan()
}

func (c *client) runCli() error {
	fmt.Println("Welcome to keychain-cli!")

	sc := bufio.NewScanner(os.Stdin)

	for prompt(sc, "%s> ", c.name) {
		line := sc.Text()

		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}

		if err := c.processLocal(parts); err != nil {
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
		cli.StringFlag{
			Name:  "datadir, D",
			Usage: "Specifies the data DIRECTORY.",
		},
	}

	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	host    string
	port    uint
	dataDir string
}

type remoteClient struct {
	conn net.Conn
	wr   *resp.Writer
	rd   *resp.Reader
}

func dial(network, address string) (*remoteClient, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	return &remoteClient{
		conn: conn,
		rd:   resp.NewReader(conn),
		wr:   resp.NewWriter(conn),
	}, nil
}

type client struct {
	cfg    *config
	local  *keychain.Keychain
	remote *remoteClient
	name   string
	mode   int // 0 for local, 1 for remote
	//remote *conn
}

func newClient(cfg *config) (*client, error) {
	client := &client{}
	if cfg == nil {
		client.cfg = &config{}
	} else {
		client.cfg = cfg
	}

	if client.cfg.dataDir != "" {
		client.name = client.cfg.dataDir
		if err := client.openLocal(); err != nil {
			return nil, err
		}
	} else {
		client.name = fmt.Sprintf("%s:%d", client.cfg.host, client.cfg.port)
		if err := client.openRemote(); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *client) openLocal() error {
	fmt.Println("opening local")
	local, err := keychain.Open(c.name)
	if err != nil {
		return err
	}

	c.local = local
	return nil
}

func (c *client) openRemote() error {
	var err error
	c.remote, err = dial("tcp", c.name)
	return err
}

type cliError string

const Ok = "OK"

const (
	WrongNumberArgs cliError = "wrong number of arguments"
	UnknownCommand  cliError = "unknown command"
)

func get(c *client, args []string) (interface{}, error) {
	value, err := c.local.Get([]byte(args[0]))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	return string(value), nil
}

func set(c *client, args []string) (interface{}, error) {
	if err := c.local.Set([]byte(args[0]), []byte(args[1])); err != nil {
		return nil, err
	}

	return Ok, nil
}

func del(c *client, args []string) (interface{}, error) {
	i := 0
	for _, key := range args {
		ok, err := c.local.Remove([]byte(key))
		if err != nil {
			return nil, err
		}
		if ok {
			i += 1
		}
	}

	return i, nil
}

func merge(c *client, args []string) (interface{}, error) {
	return Ok, nil
}

func (c *client) issueCommandLocal(args []string) (interface{}, error) {
	cmd, ok := commands[args[0]]
	if !ok {
		return UnknownCommand, nil
	}

	argc := cmd.argc
	if argc < 0 {
		if len(args[1:]) < -argc {
			return WrongNumberArgs, nil
		}
	} else if len(args[1:]) != argc {
		return WrongNumberArgs, nil
	}

	return cmd.handler(c, args[1:])
}

func (c *client) formatReplyLocal(reply interface{}) string {
	switch v := reply.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case cliError:
		return fmt.Sprintf("(error) ERR %s", v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("(integer) %d", v)
	case nil:
		return fmt.Sprintf("(nil)")
	}

	return ""
}

func (c *client) processLocal(args []string) error {
	res, err := c.issueCommandLocal(args)
	if err != nil {
		return err
	}

	reply := c.formatReplyLocal(res)
	fmt.Println(reply)

	return nil
}

func run(c *cli.Context) error {
	cfg := &config{}
	if dataDir := c.String("datadir"); dataDir != "" {
		cfg.dataDir = dataDir
	} else {
		cfg.host = c.String("host")
		cfg.port = c.Uint("port")
	}

	client, err := newClient(cfg)
	if err != nil {
		return err
	}

	return client.runCli()
}
