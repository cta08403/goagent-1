package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"
)

type Net2 interface {
	NetResolveIPAddr(network, addr string) (*net.IPAddr, error)
	NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
	TlsDialTimeout(network, address string, config *tls.Config, timeout time.Duration) (*tls.Conn, error)
	HttpClientDo(req *http.Request) (*http.Response, error)
	GetTimeout() time.Duration
	SetTimeout()
	GetAddressAlias(addr string) (alias string)
}

type Filter interface {
}

type RequestFilter interface {
	Filter
	Filter(req *http.Request) (pluginName string, pluginArgs *http.Header, err error)
}

type ResponseFilter interface {
	Filter
	Filter(req *http.Response) (newReq *http.Response, err error)
}

type PushListener interface {
	net.Listener
	Push(net.Conn, error)
}

type listenerAcceptTuple struct {
	C net.Conn
	E error
}

type listener struct {
	net.Listener
	ln net.Listener
	ch chan listenerAcceptTuple
}

func Listen(network string, addr string) (net.Listener, error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	l := &listener{
		ln: ln,
		ch: make(chan listenerAcceptTuple, 200),
	}
	// http://golang.org/src/pkg/net/http/server.go
	go func(ln net.Listener, ch chan listenerAcceptTuple) {
		var tempDelay time.Duration
		for {
			c, e := ln.Accept()
			ch <- listenerAcceptTuple{c, e}
			if e != nil {
				if ne, ok := e.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					log.Printf("http: Accept error: %v; retrying in %v", e, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				return
			}
		}
	}(l.ln, l.ch)
	return l, nil
}

func (l listener) Accept() (net.Conn, error) {
	t := <-l.ch
	return t.C, t.E
}

func (l listener) CLose() error {
	return l.ln.Close()
}

func (l listener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l listener) Push(conn net.Conn, err error) {
	l.ch <- listenerAcceptTuple{conn, err}
}

type Handler struct {
	http.Handler
	L               net.Listener
	Net             Net2
	Log             *log.Logger
	requestFilters  []RequestFilter
	responseFilters []ResponseFilter
	plugins         map[string]Plugin
}

type PluginContext struct {
	H    *Handler
	Args *http.Header
}

type Plugin interface {
	Handle(*PluginContext, http.ResponseWriter, *http.Request)
}

func (h *Handler) RegisterPlugin(name string, plugin Plugin) {
	if h.plugins == nil {
		h.plugins = make(map[string]Plugin)
	}
	h.plugins[name] = plugin
}

func (h *Handler) StackFilter(filter Filter) {
	if reqFilter, ok := filter.(RequestFilter); ok {
		h.requestFilters = append(h.requestFilters, reqFilter)
	} else if resFilter, ok := filter.(ResponseFilter); ok {
		h.responseFilters = append(h.responseFilters, resFilter)
	} else {
		h.Log.Fatalf("%s is not a valid filter", filter)
	}
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for i := len(h.requestFilters) - 1; i >= 0; i-- {
		name, args, err := h.requestFilters[i].Filter(req)
		if err != nil {
			h.Log.Fatalf("ServeHTTP error: %v", err)
		}
		if name == "" {
			continue
		}
		if plugin, ok := h.plugins[name]; ok {
			context := &PluginContext{&h, args}
			plugin.Handle(context, rw, req)
			break
		} else {
			h.Log.Fatalf("plugin \"%s\" not registered", name)
		}
	}
}
