BIN_DIR=bin
MODULE=github.com/maybetheresloop/keychain

.PHONY: all
all: keychain-server keychain-cli read

keychain-server:
	go build -o bin/$@ -v ${MODULE}/cmd/server

keychain-cli:
	go build -o bin/$@ -v ${MODULE}/cmd/cli

read:
	go build -o bin/$@ -v ${MODULE}/tools/read

.PHONY: clean test cov

clean:
	rm -rf bin/

test:
	go test ${MODULE}/...

cov:
	go test ${MODULE}/... --coverprofile coverage.txt --covermode atomic