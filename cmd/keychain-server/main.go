package main

import (
	"io"
	"net"
	"os"

	"github.com/maybetheresloop/keychain/pkg/resp"

	"github.com/maybetheresloop/keychain/internal"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const SockAddrUnix = "/var/keychain/keychain.sock"
const SockAddrTcp = ":7878"

func echoServer(conn net.Conn) {
	io.Copy(conn, conn)
	conn.Close()
}

func processConnection(conn net.Conn) {

}

type State struct {
}

func handleConnection(conn net.Conn, state *State) {
	//r := resp.NewReader(conn)
	//
	//for {
	//
	//}
}

func startUnixServer(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("accept error: %v", err)
		}

		log.Printf("accepted connection from %s\n", conn.RemoteAddr().String())

		_ = conn.Close()
	}
}

func run(c *cli.Context) error {
	fp := c.String("file")
	log.Infof("Using database file: %s", fp)

	_, err := internal.Open(fp)
	if err != nil {
		return err
	}

	log.Info("Starting server on port 7878...")

	lis, err := net.Listen("tcp", ":7878")
	if err != nil {
		return err
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Errorf("error accepting connection: %v", err)
		}

		r := resp.NewReader(conn)

		for {
			msg, err := r.ReadMessage()
			if err != nil {
				log.Errorf("error reading message: %v", err)
				continue
			}

			switch v := msg.(type) {
			case string:
				log.Infof("received command: %q", v)
			case int, int64:
				log.Infof("received integer: %d", v)
			}
		}

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

	//if err := os.RemoveAll(SockAddrUnix); err != nil {
	//	log.Panic(err)
	//}

	//lisUnix, err := net.Listen("unix", SockAddrUnix)
	//if err != nil {
	//	log.Fatalf("listen error: %v", err)
	//}

	//lisTcp, err := net.Listen("tcp", SockAddrTcp)
	//if err != nil {
	//	log.Fatalf("listen error: %v", err)
	//}

	//go startUnixServer(lisUnix)

	//for {
	//	conn, err := lisTcp.Accept()
	//	if err != nil {
	//		log.Fatalf("accept error: %v", err)
	//	}
	//
	//	log.Printf("accepted connection from %s\n", conn.RemoteAddr().String())
	//
	//	go echoServer(conn)
	//}
}
