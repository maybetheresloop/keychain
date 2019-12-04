FROM golang:1.13.4-alpine3.10

COPY bin/keychain-server /keychain-server

EXPOSE 7878

ENTRYPOINT ["/keychain-server"]