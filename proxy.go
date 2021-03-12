package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type override struct {
	Header string
	Match  string
	Host   string
	Path   string
}

type config struct {
	Path     string
	Host     string
	Override override
}

var configuration = []config{
	config{
		Path: "/helloworld-mongodb",
		Host: "kind.hostgate.net",
	},
	config{
		Path: "/mgo",
		Host: "kind.hostgate.net",
		Override: override{
			Header: "X-BF-Testing",
			Match:  "integralist",
			Path:   "/helloworld-mongodb",
		},
	},
	config{
		Path: "/independent",
		Host: "www.independent.co.uk",
		Override: override{
			Header: "X-BF-Testing",
			Match:  "integralist",
			Path:   "/",
		},
	},
	config{
		Path: "/theguardian",
		Host: "www.theguardian.com",
		Override: override{
			Header: "X-BF-Testing",
			Match:  "integralist",
			Path:   "/uk",
		},
	},
}

func generateProxy(conf config) http.Handler {
	proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
		originHost := conf.Host
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originHost)
		req.Host = originHost
		req.URL.Host = originHost
		req.URL.Scheme = "https"

		if conf.Override.Header != "" && conf.Override.Match != "" {
			if req.Header.Get(conf.Override.Header) == conf.Override.Match {
				req.URL.Path = conf.Override.Path
			}
		}
	}, Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	}}

	return proxy
}
