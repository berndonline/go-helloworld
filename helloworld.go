package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type api struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

type allContent []api

var contents = allContent{
	{
		ID:          "1",
		Name:        "Content name 1",
		Description: "Description of 1st content",
	},
	{
		ID:          "2",
		Name:        "Content name 2",
		Description: "Description of 2nd content",
	},
}

const (
	ADMIN_USER     = "admin"
	ADMIN_PASSWORD = "password"
)

func BasicAuth(handler http.HandlerFunc, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user),
			[]byte(ADMIN_USER)) != 1 || subtle.ConstantTimeCompare([]byte(pass),
			[]byte(ADMIN_PASSWORD)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("You are Unauthorized to access the application.\n"))
			return
		}
		handler(w, r)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: received a request")

	response := os.Getenv("RESPONSE")

	if response == "" {
		response = "Hello World!"
	}

	fmt.Fprintf(w, response+"\n"+os.Getenv("HOSTNAME"))
}

func createContent(w http.ResponseWriter, r *http.Request) {
	var newContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the content title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newContent)
	contents = append(contents, newContent)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newContent)
}

func getOneContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]

	for _, singleContent := range contents {
		if singleContent.ID == contentID {
			json.NewEncoder(w).Encode(singleContent)
		}
	}
}

func getAllContent(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(contents)
}

func updateContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]
	var updatedContent api

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Enter data with the content name and description only in order to update")
	}
	json.Unmarshal(reqBody, &updatedContent)

	for i, singleContent := range contents {
		if singleContent.ID == contentID {
			singleContent.Name = updatedContent.Name
			singleContent.Description = updatedContent.Description
			contents = append(contents[:i], singleContent)
			json.NewEncoder(w).Encode(singleContent)
		}
	}
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]

	for i, singleContent := range contents {
		if singleContent.ID == contentID {
			contents = append(contents[:i], contents[i+1:]...)
			fmt.Fprintf(w, "The content with ID %v has been deleted successfully", contentID)
		}
	}
}

func main() {

	log.Print("helloworld: is starting...")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", handler)
	router.HandleFunc("/api/v1/content", BasicAuth(createContent, "Please enter your username and password")).Methods("POST")
	router.HandleFunc("/api/v1/contents", BasicAuth(getAllContent, "Please enter your username and password")).Methods("GET")
	router.HandleFunc("/api/v1/content/{id}", BasicAuth(getOneContent, "Please enter your username and password")).Methods("GET")
	router.HandleFunc("/api/v1/content/{id}", BasicAuth(updateContent, "Please enter your username and password")).Methods("PATCH")
	router.HandleFunc("/api/v1/content/{id}", BasicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		log.Fatal("error starting http server : ", err)
		return
	}
}
