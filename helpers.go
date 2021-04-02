package main

import (
	"net/http"
	"strings"
)

// function to get IP address from http header - example below for custom CloudFlare X-Forwarder-For header
func getIPAddress(r *http.Request) string {
	ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
	return ip
}
