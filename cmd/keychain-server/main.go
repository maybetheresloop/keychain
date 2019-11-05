package main

import (
	"io"
	"log"
	"net"
	"os"
)

const SockAddrUnix = "/var/keychain/keychain.sock"
const SockAddrTcp = ":7878"

func echoServer(conn net.Conn) {
	io.Copy(conn, conn)
	conn.Close()
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

func main() {
	if err := os.RemoveAll(SockAddrUnix); err != nil {
		log.Panic(err)
	}

	lisUnix, err := net.Listen("unix", SockAddrUnix)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	lisTcp, err := net.Listen("tcp", SockAddrTcp)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	go startUnixServer(lisUnix)

	for {
		conn, err := lisTcp.Accept()
		if err != nil {
			log.Fatalf("accept error: %v", err)
		}

		log.Printf("accepted connection from %s\n", conn.RemoteAddr().String())

		go echoServer(conn)
	}
}
