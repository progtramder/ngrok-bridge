package ngrokbridge

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func StartTcpServer(addr string, serveHandler func(net.Conn)) error {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return err
	}

	for {
		conn, _ := l.Accept()
		if conn != nil {
			go serveHandler(conn)
		}
	}
}

func ReadRequest(c net.Conn) (*http.Request, error) {
	r := bufio.NewReader(c)
	c.SetReadDeadline(time.Now().Add(time.Second * 3))
	return http.ReadRequest(r)
}

func Start(addr string) {
	StartTcpServer(addr, func(conn net.Conn) {
		defer conn.Close()
		for {
			request, err := ReadRequest(conn)
			if err != nil {
				break
			}

			tunnel, err := GetTunnel(request.URL.Path)
			if err != nil {
				break
			}
			url := fmt.Sprintf("%s://%s%s", tunnel.Schema, tunnel.Host, request.RequestURI)
			newReq, err := http.NewRequest(request.Method, url, request.Body)
			newReq.Header = make(http.Header)
			for h, val := range request.Header {
				newReq.Header[h] = val
			}

			resp, err := tunnel.Client.Do(newReq)
			if err != nil {
				break
			}
			resp.Write(conn)
			resp.Body.Close()
		}
	})
}
