package main

import (
	"errors"
	"net"

	"github.com/maybetheresloop/keychain"
	"github.com/maybetheresloop/keychain/pkg/resp"
)

type client interface {
	Del(keys ...string) (int, error)
	Set(key string, value string) (interface{}, error)
	Get(key string) ([]byte, error)

	Close() error
}

type remoteClient struct {
	w    *resp.Writer
	rd   *resp.Reader
	conn net.Conn
}

func openRemote(addr string) (client, error) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &remoteClient{
		w:    resp.NewWriter(conn),
		conn: conn,
	}, nil
}

func (r *remoteClient) Del(keys ...string) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, "del")
	for _, key := range keys {
		args = append(args, key)
	}

	if err := r.w.WriteCommand(args...); err != nil {
		return 0, err
	}

	res, err := r.rd.ReadMessage(nil)
	if err != nil {
		return 0, err
	}

	switch v := res.(type) {
	case int64:
		return int(v), nil
	case resp.RespError:
		return 0, &v
	default:
		return 0, errors.New("unexpected response")
	}
}

func (r *remoteClient) Set(key string, value string) (interface{}, error) {
	return nil, nil
}

func (r *remoteClient) Get(key string) ([]byte, error) {
	return nil, nil
}

func (r *remoteClient) Close() error {
	return r.conn.Close()
}

type localClient struct {
	keys   *keychain.Keychain
	dbName string
}

func (l *localClient) Del(keys ...string) (int, error) {

	removed := 0
	for _, key := range keys {
		ok, err := l.keys.Remove([]byte(key))
		if err != nil {
			return removed, err
		}

		if ok {
			removed += 1
		}
	}

	return removed, nil
}

func (l *localClient) Set(key string, value string) (interface{}, error) {
	err := l.keys.Set([]byte(key), []byte(value))
	if err != nil {
		return nil, err
	}

	return "OK", nil
}

func (l *localClient) Get(key string) ([]byte, error) {
	return l.keys.Get([]byte(key))
}

func (l *localClient) Close() error {
	return l.keys.Close()
}
