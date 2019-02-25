package ngrokbridge

import (
	"bytes"
	"io"
	"net"
)

const (
	initBufSize = 1024 // allocate 1 KB up front to try to avoid resizing
)

type Conn struct {
	net.Conn               // the raw connection
	buf *bytes.Buffer // all of the initial data that has to be read in order to vhost a connection is saved here
}

func NewConn(conn net.Conn) (*Conn, io.Reader) {
	c := &Conn{
		Conn:     conn,
		buf: bytes.NewBuffer(make([]byte, 0, initBufSize)),
	}

	return c, io.TeeReader(conn, c.buf)
}

func (c *Conn) Read(p []byte) (n int, err error) {
	if c.buf == nil {
		return c.Conn.Read(p)
	}
	n, err = c.buf.Read(p)

	// end of the request buffer
	if err == io.EOF {
		// let the request buffer get garbage collected
		// and make sure we don't read from it again
		c.buf = nil

		// continue reading from the connection
		var n2 int
		n2, err = c.Conn.Read(p[n:])

		// update total read
		n += n2
	}
	return
}
