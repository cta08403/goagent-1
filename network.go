package main

import (
	"crypto/tls"
	// "fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type SimpleNetwork struct {
	Net2
}

func (sn *SimpleNetwork) NetResolveIPAddr(network, addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr(network, addr)
}

func (sn *SimpleNetwork) NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func (sn *SimpleNetwork) TlsDialTimeout(network string, addr string, config *tls.Config, timeout time.Duration) (*tls.Conn, error) {
	return tls.Dial(network, addr, config)
}

func (sn *SimpleNetwork) HttpClientDo(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func (sn *SimpleNetwork) CopyResponseBody(w io.Writer, res *http.Response) (int64, error) {
	return io.Copy(w, res.Body)
}

func (sn *SimpleNetwork) GetTimeout() time.Duration {
	return 8 * time.Second
}

func (sn *SimpleNetwork) SetTimeout() {
}

func (sn *SimpleNetwork) GetAddressAlias(addr string) (alias string) {
	return ""
}

type AdvancedNetwork struct {
	Net2
	dnsCache   map[string]*net.IPAddr
	dnsCacheMu sync.Mutex
}

func NewAdvancedNetwork() *AdvancedNetwork {
	return &AdvancedNetwork{
		dnsCache: map[string]*net.IPAddr{},
	}
}

func (an *AdvancedNetwork) NetResolveIPAddr(network, addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr(network, addr)
}

func (an *AdvancedNetwork) NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func (an *AdvancedNetwork) TlsDialTimeout(network string, addr string, config *tls.Config, timeout time.Duration) (*tls.Conn, error) {
	return tls.Dial(network, addr, config)
}

func (an *AdvancedNetwork) HttpClientDo(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func (an *AdvancedNetwork) GetTimeout() time.Duration {
	return 8 * time.Second
}

func (s *AdvancedNetwork) SetTimeout() {
}

func (s *AdvancedNetwork) GetAddressAlias(addr string) (alias string) {
	return ""
}
