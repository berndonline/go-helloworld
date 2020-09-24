package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: received a request")
	response := os.Getenv("RESPONSE")
	if response == "" {
		response = "Hello World!"
	}
	fmt.Fprintf(w, response + "\n" + os.Getenv("HOSTNAME"))
}

func main() {
	log.Print("helloworld application is starting...")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
