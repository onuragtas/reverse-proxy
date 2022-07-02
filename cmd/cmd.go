package main

import (
	"fmt"
	proxy2 "github.com/onuragtas/reverse-proxy/proxy"
	"log"
	"net"
	"strings"
	"time"
)

var (
	localAddr  = "0.0.0.0:8889"
	remoteAddr = "localhost:80"
)

func main() {

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		panic("connection error:" + err.Error())
	}
	log.Println("Proxy listening on", localAddr, "...")
	for {
		conn, err := listener.Accept()
		proxy := proxy2.Proxy{Src: conn, OnResponse: onResponse, OnRequest: onRequest, RequestDestination: setDestination, RequestHost: setHost}
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		go proxy.Handle()
	}
}

func onRequest(srcLocal, srcRemote, dstLocal, dstRemote string, request []byte) {
	log.Println(srcLocal, "->", srcRemote, "->", dstLocal, "->", dstRemote, string(request))
}

func onResponse(dstRemote, dstLocal, srcRemote, srcLocal string, response []byte) {
	log.Println(dstRemote, "->", dstLocal, "->", srcRemote, "->", srcLocal)
}

func setDestination(host string) net.Conn {
	rHost := ""
	split := strings.Split(host, ":")
	if split[0] == "localhost" {
		rHost = "onur.resoft.org:80"
	}
	rHost = remoteAddr

	var err error
	tcp, err := net.DialTimeout("tcp", rHost, time.Second*10)
	if err != nil {
		log.Println(err)
	}
	return tcp
}

func setHost(host string) string {
	split := strings.Split(host, ":")
	if split[0] == "localhost" {
		return "onur.resoft.org:88"
	}
	return remoteAddr
}
