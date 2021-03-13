package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type override struct {
	Path   string
	User   string
	Pass   string
	Scheme string
}

type config struct {
	Path     string
	Host     string
	Override override
}

var configuration = []config{
	config{
		Path: "/hello/mgo",
		Host: "helloworld-mongodb.helloworld.svc.cluster.local",
		Override: override{
			Path:   "/api/v1/content/",
			User:   "user1",
			Pass:   "password1",
			Scheme: "http",
		},
	},
	config{
		Path: "/hello",
		Host: "helloworld.helloworld.svc.cluster.local",
		Override: override{
			Path:   "/api/v1/content/",
			User:   "user1",
			Pass:   "password1",
			Scheme: "http",
		},
	},
	config{
		Path: "/independent",
		Host: "www.independent.co.uk",
		Override: override{
			Path: "/",
		},
	},
	config{
		Path: "/theguardian",
		Host: "www.theguardian.com",
		Override: override{
			Path: "/uk",
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

		if conf.Override.Path != "" {
			req.URL.Path = conf.Override.Path

			if conf.Override.User != "" {
				req.SetBasicAuth(conf.Override.User, conf.Override.Pass)
			}
			if conf.Override.Scheme != "" {
				req.URL.Scheme = conf.Override.Scheme
			}
		}

	}, Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	}}

	return proxy
}
