package main

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type config struct {
	Path     string
	Host     string
	Override override
}

type override struct {
	Path   string
	User   string
	Pass   string
	Scheme string
}

var configuration = []config{
	config{
		Path: "/proxy",
		Host: "helloworld.helloworld.svc.cluster.local",
		Override: override{
			Path:   "/api/v1/content",
			User:   "user1",
			Pass:   "password1",
			Scheme: "http",
		},
	},
	config{
		Path: "/proxy/mgo",
		Host: "helloworld-mongodb.helloworld.svc.cluster.local",
		Override: override{
			Path:   "/api/v1/content",
			User:   "user1",
			Pass:   "password1",
			Scheme: "http",
		},
	},
	config{
		Path: "/proxy/independent",
		Host: "www.independent.co.uk",
		Override: override{
			Path: "/",
		},
	},
	config{
		Path: "/proxy/theguardian",
		Host: "www.theguardian.com",
		Override: override{
			Path: "/uk",
		},
	},
}

func generateProxy(conf config) http.Handler {
	proxy := &httputil.ReverseProxy{Director: func(r *http.Request) {
		originHost := conf.Host

		r.Header.Add("X-Forwarded-Host", r.Host)
		r.Header.Add("X-Origin-Host", originHost)
		r.Host = originHost
		r.URL.Host = originHost
		r.URL.Scheme = "https"

		if conf.Override.Path != "" {
			r.URL.Path = conf.Override.Path

			if conf.Override.User != "" {
				r.SetBasicAuth(conf.Override.User, conf.Override.Pass)
			}
			if conf.Override.Scheme != "" {
				r.URL.Scheme = conf.Override.Scheme
			}
		}

	}, Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	}}

	return proxy
}
