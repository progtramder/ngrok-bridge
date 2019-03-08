package ngrokbridge

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net"
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

//Peek the request path
func readPath(r *bufio.Reader) ([]byte, error) {
	b, _, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	i := bytes.Index(b, []byte("/"))
	if i == -1 {
		return nil, errors.New("malformed http request")
	}

	for j := i + 1; j < len(b); j++ {
		if b[j] == '?' || b[j] == ' ' {
			return b[i : j], nil
		}
	}

	return nil, errors.New("malformed http request")
}

func handler(conn net.Conn) {
	defer conn.Close()
	//Wrap the conn with TeeReader
	c, r := NewConn(conn)
	c.SetReadDeadline(time.Now().Add(time.Second * 2))
	path, err := readPath(bufio.NewReader(r))
	if err != nil {
		return
	}
	tunnel, err := GetTunnel(string(path))
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
