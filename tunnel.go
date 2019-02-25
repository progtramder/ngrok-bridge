package ngrokbridge

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)


var tunnels = map[string]*Tunnel{}
var httpClient = &http.Client{}
var tlsClient  = &http.Client{
	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
}

type Tunnel struct {
	Schema string
	Host   string
	Client *http.Client
}

func MakeTunnel(configFile string) error {

	ioutil.ReadFile(configFile)
	//To be implemented
	return nil
}

func RegisterTunnel(schema, host string, paths []string) {
	var client *http.Client
	if schema == "http" {
		client = httpClient
	} else if schema == "https"{
		client = tlsClient
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
