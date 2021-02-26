package main

import (
  "crypto/subtle"
	"log"
	"net/http"
)

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user),
			[]byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass),
			[]byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("You are Unauthorized to access the application.\n"))
			log.Print("helloworld-api: authentication failed - " + getIPAddress(r))
			return
		}
		handler(w, r)
	}
}
