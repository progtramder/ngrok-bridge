package ngrokbridge

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type TcpHandler interface {
	Handle(net.Conn)
}

type TcpFuncHandler func(net.Conn)

func (self TcpFuncHandler) Handle(conn net.Conn) {
	self(conn)
}

func StartTcpServer(addr string, serveHandler TcpHandler) error {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, _ := l.Accept()
		if conn != nil {
			go serveHandler.Handle(conn)
		}
	}
}

func handler(conn net.Conn) {
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
}

func Start(addr string) error {
	return StartTcpServer(addr, TcpFuncHandler(handler))
}

func StartTLS(addr string) error {
	return StartTcpServer(addr, TcpFuncHandler(func(conn net.Conn) {
		config := &tls.Config{}
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0] = rootCert
		//Set the connection as TLS connection
		tlsConn := tls.Server(conn, config)
		handler(tlsConn)
	}))
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
