package main

import (
	"crypto/tls"
	"io"
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
	CopyResponseBody(w io.Writer, res *http.Response) (int64, error)
	GetTimeout() time.Duration
	SetTimeout()
	GetAddressAlias(addr string) (alias string)
}

type RequestFilter interface {
	Filter(req *http.Request) (pluginName string, pluginArgs *http.Header, err error)
}

type ResponseFilter interface {
	Filter(req *http.Request, res *http.Response) (pluginName string, pluginArgs *http.Header, err error)
}

type PushListener interface {
	net.Listener
	Push(net.Conn, error)
}

type listenerAcceptTuple struct {
	c net.Conn
	e error
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
	return t.c, t.e
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
	Listener        net.Listener
	Log             *log.Logger
	Net             Net2
	RequestPlugins  map[string]RequestPlugin
	ResponsePlugins map[string]ResponsePlugin
	RequestFilters  []RequestFilter
	ResponseFilters []ResponseFilter
}

type PluginContext struct {
	H    *Handler
	Args *http.Header
}

type RequestPlugin interface {
	HandleRequest(*PluginContext, http.ResponseWriter, *http.Request) (*http.Response, error)
}

type ResponsePlugin interface {
	HandleResponse(*PluginContext, http.ResponseWriter, *http.Request, *http.Response, error) error
}

type Plugin interface {
	RequestPlugin
	ResponsePlugin
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, reqfilter := range h.RequestFilters {
		name, args, err := reqfilter.Filter(req)
		if err != nil {
			h.Log.Printf("RequestFilter error: %v", err)
		}
		if name == "" {
			continue
		}
		if reqplugin, ok := h.RequestPlugins[name]; ok {
			reqctx := &PluginContext{&h, args}
			res, err := reqplugin.HandleRequest(reqctx, rw, req)
			if err != nil {
				h.Log.Printf("Plugin %s HandleResponse error: %v", name, err)
			}
			if res != nil {
				for _, resfilter := range h.ResponseFilters {
					name, args, err := resfilter.Filter(req, res)
					if err != nil {
						h.Log.Printf("ServeHTTP RequestFilter error: %v", err)
					}
					if name == "" {
						continue
					}
					if resplugin, ok := h.ResponsePlugins[name]; ok {
						resctx := &PluginContext{&h, args}
						err := resplugin.HandleResponse(resctx, rw, req, res, err)
						if err != nil {
							h.Log.Printf("Plugin %s HandleResponse error: %v", name, err)
						}
					}
					break
				}
			}
			break
		} else {
			h.Log.Fatalf("plugin \"%s\" not registered", name)
		}
	}
}
