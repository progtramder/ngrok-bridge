package ngrokbridge

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
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

func Start(addr string) {
	StartTcpServer(addr, func(conn net.Conn) {
		defer conn.Close()
		//Wrap the conn with TeeReader
		c, r := NewConn(conn)
		c.SetReadDeadline(time.Now().Add(time.Second * 2))
		request, err := http.ReadRequest(bufio.NewReader(r))
		if err != nil {
			return
		}
		request.Body.Close()

		tunnel, err := GetTunnel(request.URL.Path)
		if err != nil {
			return
		}
		cServer, err := tunnel.GetProxy()
		if err != nil {
			return
		}

		c.SetDeadline(time.Time{})
		Join(cServer, c)
	})
}

//Copy data between two conn in full duplex mode,
//and quit once a conn occur an error
func Join(c, c2 net.Conn) {
	var wait sync.WaitGroup

	pipe := func(to, from net.Conn) {
		defer to.Close()
		defer from.Close()
		defer wait.Done()
		io.Copy(to, from)
	}

	wait.Add(2)
	go pipe(c, c2)
	go pipe(c2, c)
	wait.Wait()
}
