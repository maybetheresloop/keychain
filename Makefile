BIN_DIR=bin
MODULE=github.com/maybetheresloop/keychain

.PHONY: all
all: keychain keychain-server create-db read

keychain:
	go build -o bin/$@ -v ${MODULE}/cmd/keychain

keychain-server:
	go build -o bin/$@ -v ${MODULE}/cmd/keychain-server

create-db:
	go build -o bin/$@ -v ${MODULE}/tools/create-db

read:
	go build -o bin/$@ -v ${MODULE}/tools/read

.PHONY: clean test

clean:
	rm -rf bin/

test:
	go test ${MODULE}/...