package proxy

import (
	"fmt"
	"net"
	"testing"
)

var (
	wafPort  = "0.0.0.0:8889"
	realPort = "localhost:3333"
)

func TestNewProxy(t *testing.T) {

	listener, err := net.Listen("tcp", wafPort)
	if err != nil {
		t.Errorf("connection error:" + err.Error())
	}

	for {
		conn, err := listener.Accept()
		proxy := Proxy{Src: conn, Destination: realPort}
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		go proxy.Handle()
	}
}
