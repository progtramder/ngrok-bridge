package ngrokbridge

import (
	"crypto/tls"
	"errors"
	"gopkg.in/yaml.v2"
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
	type tun struct {
		Schema string   `yaml:"schema"`
		Host   string   `yaml:"host"`
		Path   []string `yaml:"path"`
	}

	router := struct {
		Router []tun `yaml:"router"`
	}{}

	setting, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(setting, &router)
	if err != nil {
		return err
	}

	for _, tun := range router.Router {
		RegisterTunnel(tun.Schema, tun.Host, tun.Path)
	}

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
			panic("tunnel path conflict : " + t)
		}
		tunnels[t] = tunnel
	}
}

func GetTunnel(path string) (*Tunnel, error) {
	t, ok := tunnels[path]
	if ok {
		return t, nil
	}

	var (
		index = -1
		key string
	)

	//find best match. for path=/service/sub/test, registered path /service/sub
	//is better than /service and even better than /, that means the math index
	// the higher the better
	for k := range tunnels {
		i := strings.Index(path, k)
		if i > index {
			index = i
			key = k
		}
	}

	if index == -1 {
		return nil, errors.New("tunnel not found : " + path)
	}

	return tunnels[key], nil
}
