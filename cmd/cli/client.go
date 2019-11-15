package main

import (
	"github.com/maybetheresloop/keychain"
)

type client interface {
	Del(keys ...string) (int, error)
	Set(key string, value string) (interface{}, error)
	Get(key string) ([]byte, error)

	Close() error
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
