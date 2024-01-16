package proxy

import (
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

type Proxy struct {
	Src                net.Conn
	destination        net.Conn
	Destination        string
	OnRequest          func(srcLocal, srcRemote, dstLocal, dstRemote string, request []byte)
	OnResponse         func(dstRemote, dstLocal, srcRemote, srcLocal string, response []byte)
	RequestDestination func(host string) net.Conn
	RequestHost        func(request []byte, host string, src net.Conn) string
	OnCloseSource      func(conn net.Conn)
	OnCloseDestination func(conn net.Conn)
}

func (t *Proxy) Handle() {
	defer func() {
		if t.OnCloseSource != nil {
			t.OnCloseSource(t.Src)
		}
	}()

	defer func() {
		if t.OnCloseDestination != nil {
			t.OnCloseDestination(t.destination)
		}
	}()

	srcCloseChan := make(chan bool)
	dstCloseChan := make(chan bool)

	readFromSrcChan := make(chan []byte)
	readFromDstChan := make(chan []byte)

	go func() {
		for true {
			request := <-readFromSrcChan
			host := t.getHostIfHttp(request)
			if host != "" {
				host = t.RequestHost(request, host, t.Src)
				request = t.changeRequest(request, host)
			}
			if t.destination != nil {
				_, err := t.destination.Write(request)
				if err != nil {
					return
				}
				if err != nil {
					log.Println(err)
				}
				if t.OnRequest != nil {
					t.OnRequest(t.Src.LocalAddr().String(), t.Src.RemoteAddr().String(), t.destination.LocalAddr().String(), t.destination.RemoteAddr().String(), request)
				}
			} else {
				dest := t.RequestHost(request, host, t.Src)
				t.Destination = dest
				t.DestinationConnect()
				if t.destination != nil {
					_, err := t.destination.Write(request)
					if err != nil {
						return
					}
					if err != nil {
						log.Println(err)
					}
					if t.OnRequest != nil {
						t.OnRequest(t.Src.LocalAddr().String(), t.Src.RemoteAddr().String(), t.destination.LocalAddr().String(), t.destination.RemoteAddr().String(), request)
					}
				}
			}
			if t.destination == nil {
				t.Src.Close()
				srcCloseChan <- true
				dstCloseChan <- true
			}
		}
	}()
	go func() {
		for {
			response := <-readFromDstChan
			_, err := t.Src.Write(response)
			if err != nil {
				log.Println(err)
			}
			if t.OnResponse != nil {
				t.OnResponse(t.destination.RemoteAddr().String(), t.destination.LocalAddr().String(), t.Src.RemoteAddr().String(), t.Src.LocalAddr().String(), response)
			}
		}
	}()

	go func() {
		for {
			if t.destination != nil {
				err := t.destination.SetDeadline(time.Now().Add(100 * time.Second))
				buf := make([]byte, 8192)
				n, err := t.destination.Read(buf)
				readFromDst := buf[:n]
				readFromDstChan <- readFromDst
				if err != nil {
					dstCloseChan <- true
					break
				}
			} else {
				t.Src.Close()
				srcCloseChan <- true
				dstCloseChan <- true
			}
		}
	}()

	go func() {
		for {
			err := t.Src.SetDeadline(time.Now().Add(100 * time.Second))
			if err != nil {
				log.Println(err)
			}

			buf := make([]byte, 8192)
			n, err := t.Src.Read(buf)
			readFromSrc := buf[:n]

			if t.Destination == "" {
				host := t.getHostIfHttp(readFromSrc)
				if host != "" {
					t.destination = t.RequestDestination(host)
				}
				if err != nil {
					log.Println(err)
				}
			}
			readFromSrcChan <- readFromSrc
			if err != nil {
				srcCloseChan <- true
				dstCloseChan <- true
				break
			}
		}
	}()

	<-srcCloseChan
	<-dstCloseChan
}

func (t *Proxy) changeRequest(request []byte, destination string) []byte {
	if strings.Contains(string(request), "HTTP/1.1") {
		var re = regexp.MustCompile(`(?m)Host: ([A-Za-z0-9-_:.]+)`)
		request = []byte(re.ReplaceAllString(string(request), "Host: "+destination))
	}

	return request
}

func (t *Proxy) getHostIfHttp(request []byte) string {
	if strings.Contains(string(request), "HTTP/1.1") {
		var re = regexp.MustCompile(`(?m)Host: ([A-Za-z0-9-_:.]+)`)
		return strings.ReplaceAll(re.FindString(string(request)), "Host: ", "")
	}

	return ""
}

func (t *Proxy) DestinationConnect() {
	var err error
	t.destination, err = net.DialTimeout("tcp", t.Destination, time.Second*10)
	if err != nil {
		log.Println(err)
	}
}
