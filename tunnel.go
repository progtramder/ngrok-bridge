package ngrokbridge

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
)


var tunnels = map[string]*Tunnel{}

type Tunnel struct {
	Schema string
	Host   string
	Client *http.Client
}

func MakeTunnel(configFile string) {

}

func RegisterTunnel(schema, host string, paths []string) {
	var client *http.Client
	if schema == "http" {
		client = &http.Client{}
	} else if schema == "https"{
		tp := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport:tp}
	} else {
		panic("invalid schema")
	}

	tunnel := &Tunnel{schema, host, client}
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
