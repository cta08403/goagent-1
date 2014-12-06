package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type StripRequestPlugin struct {
	RequestPlugin
}

func (p StripRequestPlugin) HandleRequest(c *PluginContext, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, errors.New("http.ResponseWriter does not implments Hijacker")
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("http.ResponseWriter Hijack failed: %s", err))
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	c.H.Log.Printf("%s \"STRIP %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
	cert, err := tls.LoadX509KeyPair("./certs/.google.com.crt", "./certs/.google.com.crt")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("tls.LoadX509KeyPair failed: %s", err))
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.VerifyClientCertIfGiven}
	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, errors.New(fmt.Sprintf("tlsConn.Handshake error: %s", err))
	}
	if pl, ok := c.H.Listener.(PushListener); ok {
		pl.Push(tlsConn, nil)
		return nil, nil
	}
	loConn, err := net.Dial("tcp", c.H.Listener.Addr().String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("net.Dial failed: %s", err))
	}
	go io.Copy(loConn, tlsConn)
	go io.Copy(tlsConn, loConn)
	return nil, nil
}
