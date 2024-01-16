package proxy

import (
	"fmt"
	"net"
	"testing"
)

var (
	wafPort = "0.0.0.0:10000"
)

func TestNewProxy(t *testing.T) {

	listener, err := net.Listen("tcp", wafPort)
	if err != nil {
		t.Errorf("connection error:" + err.Error())
	}

	for {
		conn, err := listener.Accept()
		proxy := Proxy{Src: conn, RequestHost: setDestination}
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		go proxy.Handle()
	}
}

func setDestination(req []byte, host string, src net.Conn) string {
	return "127.0.0.1:9998"
}
