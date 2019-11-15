BIN_DIR=bin
MODULE=github.com/maybetheresloop/keychain

.PHONY: all
all: keychain-server keychain-cli create-db read

keychain-server:
	go build -o bin/$@ -v ${MODULE}/cmd/server

keychain-cli:
	go build -o bin/$@ -v ${MODULE}/cmd/cli

create-db:
	go build -o bin/$@ -v ${MODULE}/tools/create-db

read:
	go build -o bin/$@ -v ${MODULE}/tools/read

.PHONY: clean test cov

clean:
	rm -rf bin/

test:
	go test ${MODULE}/...

cov:
	go test ${MODULE}/... --coverprofile coverage.txt --covermode atomic