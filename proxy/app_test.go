package proxy

import (
	"fmt"
	"log"
	"net"
	"strings"
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
		proxy := Proxy{
			Src:         conn,
			RequestHost: setDestination,
			OnRequest: func(srcLocal, srcRemote, dstLocal, dstRemote string, request []byte, srcConnection net.Conn, dstConnection net.Conn) {
				log.Println("Request:", string(request))
				if strings.Contains(string(request), "tatus=\"stopping\"") {
					srcConnection.Close()
					dstConnection.Close()
				}
			},
			OnResponse: func(dstRemote, dstLocal, srcRemote, srcLocal string, response []byte, srcConnection net.Conn, dstConnection net.Conn) {
				log.Println("Response:", string(response))
			},
			OnCloseSource: func(conn net.Conn) {
				log.Println("Connection closed from", conn.RemoteAddr().String(), conn.LocalAddr().String())
			},
			OnCloseDestination: func(conn net.Conn) {
				log.Println("Connection closed to", conn.RemoteAddr().String(), conn.LocalAddr().String())
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
	return "127.0.0.1:9998"
}
