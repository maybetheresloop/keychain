BIN_DIR=bin
MODULE=github.com/maybetheresloop/keychain

.PHONY: all
all: keychain keychain-server

keychain:
	go build -o bin/$@ -v ${MODULE}/cmd/keychain

keychain-server:
	go build -o bin/$@ -v ${MODULE}/cmd/keychain-server

.PHONY: clean test

clean:
	rm -rf bin/

test:
	go test ${MODULE}/...