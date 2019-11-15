package main

import (
	"bytes"
	"net"
	"os"

	"github.com/maybetheresloop/keychain/pkg/resp"

	"github.com/maybetheresloop/keychain"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const SockAddrUnix = "/var/keychain/keychain.sock"
const SockAddrTcp = ":7878"

func ParseBulkStringArray(r *resp.Reader, num int64) (interface{}, error) {

	return nil, nil
}

// Processes a client connection. This can be a connection through either a TCP socket or a Unix domain socket.
func processConnection(conn net.Conn, state *State) error {

	r := resp.NewReader(conn)

	res, err := r.ReadMessage(resp.BulkStringSliceParser)
	if err != nil {
		return err
	}

	value, ok := res.([][]byte)
	if !ok {
		return nil
	}

	if bytes.Compare(value[0], []byte("get")) == 0 {
		state.keys.Get
	}

	return nil

	//r := resp.NewReader(conn)
	//w := resp
	//
	//for {
	//	message, err := r.ReadMessage()
	//
	//	// RESP parsing errors are fatal and cause the connection to be closed immediately.
	//	if err != nil {
	//		break
	//	}
	//
	//}
}

type State struct {
	keys *keychain.Keychain
}

func run(c *cli.Context) error {
	fp := c.String("file")
	log.Infof("Using database file: %s", fp)

	_, err := keychain.Open(fp)
	if err != nil {
		return err
	}

	log.Info("Starting server on port 7878...")

	lis, err := net.Listen("tcp", ":7878")
	if err != nil {
		return err
	}

	keys, err := keychain.Open("")
	if err != nil {
		log.Errorf("failed to open database file: %v", err)
	}

	state := &State{
		keys: keys,
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Errorf("error accepting connection: %v", err)
		}

		go processConnection(conn, state)
	}

	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "keychain-server"
	app.Usage = "Start, stop, or restart a Keychain server."
	app.Version = "0.1.0"
	app.Action = run

	fileFlag := cli.StringFlag{
		Name:      "file, f",
		Required:  true,
		Usage:     "Database FILE to use",
		TakesFile: true,
	}

	app.Flags = []cli.Flag{
		fileFlag,
	}

	if err := app.Run(os.Args); err != nil {
		log.Info(err)
	}

}
