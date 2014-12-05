package main

import (
	"net/http"
)

type DirectFilter struct {
	RequestFilter
}

func (d *DirectFilter) Filter(req *http.Request) (pluginName string, pluginArgs *http.Header, err error) {
	return "direct", nil, nil
}

type StripFilter struct {
	RequestFilter
}

func (d *StripFilter) Filter(req *http.Request) (pluginName string, pluginArgs *http.Header, err error) {
	if req.Method == "CONNECT" {
		args := http.Header{
			"Foo": []string{"bar"},
			"key": []string{"value"},
		}
		return "strip", &args, nil
	}
	return "", nil, nil
}
