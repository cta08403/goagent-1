package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type SimpleNetwork struct {
	Net2
}

func (s *SimpleNetwork) NetResolveIPAddr(network, addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr(network, addr)
}

func (s *SimpleNetwork) NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func (s *SimpleNetwork) TlsDialTimeout(network string, addr string, config *tls.Config, timeout time.Duration) (*tls.Conn, error) {
	return tls.Dial(network, addr, config)
}

func (s *SimpleNetwork) HttpClientDo(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func (s *SimpleNetwork) GetTimeout() time.Duration {
	return 8 * time.Second
}

func (s *SimpleNetwork) SetTimeout() {
}

func (s *SimpleNetwork) GetAddressAlias(addr string) (alias string) {
	return ""
}
