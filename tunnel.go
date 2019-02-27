package ngrokbridge

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

var tunnels = map[string]*Tunnel{}

type Tunnel struct {
	Schema string
	Host   string
}

func (t *Tunnel) newConn() (net.Conn, error){
	conn, err := net.DialTimeout("tcp", t.Host, time.Second*3)
	if err != nil {
		return nil, err
	}

	if t.Schema == "https" {
		conn = tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
	}

	return conn, nil
}

func (t *Tunnel) GetProxy() (net.Conn, error) {
	return t.newConn()
}

func MakeTunnel(configFile string) error {
	ioutil.ReadFile(configFile)
	//To be implemented
	return nil
}

func RegisterTunnel(schema, host string, paths []string) {
	if !strings.Contains(host, ":") {
		if schema == "http" {
			host += ":80"
		} else if schema == "https"{
			host += ":443"
		} else {
			panic("invalid schema")
		}
	}

	tunnel := &Tunnel{schema, host}
	for _, t := range paths {
		if _, ok := tunnels[t]; ok {
			panic("tunnel name conflict : " + t)
		}
		tunnels[t] = tunnel
	}
}

func GetTunnel(path string) (*Tunnel, error) {
	t, ok := tunnels[path]
	if ok {
		return t, nil
	}

	for k, v := range tunnels {
		if strings.Contains(path, k) {
			t = v
			return t, nil
		}
	}
	return nil, errors.New("tunnel not found : " + path)
}
