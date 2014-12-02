package main

import (
	"fmt"
	"io"
	"net/http"
)

type DirectPlugin struct {
	Plugin
}

func (p DirectPlugin) Handle(c *PluginContext, rw http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		if !req.URL.IsAbs() {
			if req.TLS != nil {
				req.URL.Scheme = "https"
				if req.Host != "" {
					req.URL.Host = req.Host
				} else {
					req.URL.Host = req.TLS.ServerName
				}
			} else {
				req.URL.Scheme = "http"
				req.URL.Host = req.Host
			}
		}
		newReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
		if err != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", err)
			return
		}
		newReq.Header = req.Header
		res, err := c.H.Net.HttpClientDo(newReq)
		if err != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", err)
			return
		}
		c.H.Log.Printf("%s \"DIRECT %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
		rw.WriteHeader(res.StatusCode)
		for key, values := range res.Header {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}
		io.Copy(rw, res.Body)
	} else {
		c.H.Log.Printf("%s \"DIRECT %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
		remoteConn, err := c.H.Net.NetDialTimeout("tcp", req.Host, c.H.Net.GetTimeout())
		if err != nil {
			c.H.Log.Printf("NetDialTimeout %s failed %s", req.Host, err)
			return
		}
		hijacker, ok := rw.(http.Hijacker)
		if !ok {
			c.H.Log.Printf("http.ResponseWriter does not implments Hijacker")
			return
		}
		localConn, _, err := hijacker.Hijack()
		localConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		go io.Copy(remoteConn, localConn)
		io.Copy(localConn, remoteConn)
	}
}
