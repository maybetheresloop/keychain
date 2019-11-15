package main

type Command struct {
	handler func(c client) (interface{}, error)
}

var RunLocal = Command{}
