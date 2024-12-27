package proxy

import (
	"fmt"
	"net"
	"testing"
	"time"
)

var (
	wafPort = "0.0.0.0:1113"
)

func TestNewProxy(t *testing.T) {

	listener, err := net.Listen("tcp", wafPort)
	if err != nil {
		t.Errorf("connection error:" + err.Error())
	}

	for {
		conn, err := listener.Accept()
		proxy := Proxy{
			Timeout: time.Duration(30),
			Src:     conn,
			OnResponse: func(dstRemote, dstLocal, srcRemote, srcLocal string, response []byte, srcConnection, dstConnection net.Conn) {
				// srcConnection.Write(response)
				// dstConnection.Write(response)
			},
			OnRequest: func(srcLocal, srcRemote, dstLocal, dstRemote string, request []byte, srcConnection, dstConnection net.Conn) {
				// srcConnection.Write(request)
				// dstConnection.Write(request)
			},
			RequestHost: func(request []byte, host string, src net.Conn) string {
				return host
			},
			RequestTCPDestination: func(request []byte, host string, src net.Conn) net.Conn {
				dest, _ := net.DialTimeout("tcp", "127.0.0.1:3306", time.Second*10)
				return dest
			},
		}
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		go proxy.Handle()
	}
}

func setDestination(req []byte, host string, src net.Conn) string {
	return "api.dev.net:80"
}
