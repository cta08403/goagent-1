package main

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
)

type StripPlugin struct {
	Plugin
}

func (p StripPlugin) Handle(c *PluginContext, rw http.ResponseWriter, req *http.Request) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		c.H.Log.Printf("http.ResponseWriter does not implments Hijacker")
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		c.H.Log.Printf("http.ResponseWriter Hijack failed: %s", err)
		return
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	c.H.Log.Printf("%s \"STRIP %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
	cert, err := tls.LoadX509KeyPair("./certs/.google.com.crt", "./certs/.google.com.crt")
	if err != nil {
		c.H.Log.Printf("tls.LoadX509KeyPair failed: %s", err)
		return
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.VerifyClientCertIfGiven}
	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		c.H.Log.Printf("tlsConn.Handshake error: %s", err)
	}
	if pl, ok := c.H.L.(PushListener); ok {
		pl.Push(tlsConn, nil)
		return
	}
	loConn, err := net.Dial("tcp", c.H.L.Addr().String())
	if err != nil {
		c.H.Log.Printf("net.Dial failed: %s", err)
		return
	}
	go io.Copy(loConn, tlsConn)
	io.Copy(tlsConn, loConn)
}
