package main

import "errors"

var ErrKeyNotSpecified = errors.New("must specify a key")
var ErrValueNotSpecified = errors.New("must specify a value")
