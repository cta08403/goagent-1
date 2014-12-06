package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type DirectRequestPlugin struct {
	RequestPlugin
}

type DirectResponsePlugin struct {
	ResponsePlugin
}

func (p DirectRequestPlugin) HandleRequest(c *PluginContext, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
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
			return nil, err
		}
		newReq.Header = req.Header
		res, err := c.H.Net.HttpClientDo(newReq)
		return res, err
	} else {
		c.H.Log.Printf("%s \"DIRECT %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
		response := &http.Response{
			StatusCode:    200,
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{},
			ContentLength: -1,
		}
		return response, nil
	}
}

func (p DirectResponsePlugin) HandleResponse(c *PluginContext, rw http.ResponseWriter, req *http.Request, res *http.Response, resError error) error {
	if req.Method != "CONNECT" {
		if resError != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			return resError
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
		if resError != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			c.H.Log.Printf("NetDialTimeout %s failed %s", req.Host, resError)
			return resError
		}
		remoteConn, err := c.H.Net.NetDialTimeout("tcp", req.Host, c.H.Net.GetTimeout())
		if err != nil {
			return err
		}
		hijacker, ok := rw.(http.Hijacker)
		if !ok {
			resError = errors.New("http.ResponseWriter does not implments Hijacker")
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			return resError
		}
		localConn, _, err := hijacker.Hijack()
		localConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		go io.Copy(remoteConn, localConn)
		io.Copy(localConn, remoteConn)
	}
	return nil
}
