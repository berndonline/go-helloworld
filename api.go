package main

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
)

type api struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
}

type allContent []api

var contents = allContent{
	{
		ID:          "1",
		Name:        "Content 1",
	},
	{
		ID:          "2",
		Name:        "Content 2",
	},
}

func getIndexContent(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld-api: getIndexContent received a request")
	respondWithJson(w, http.StatusOK, contents)
}

func getSingleContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]
	for _, singleContent := range contents {
    if singleContent.ID == contentID {
			log.Print("helloworld-api: getSingleContent received a request")
			respondWithJson(w, http.StatusOK, singleContent)
      return
		}
	}
	log.Print("helloworld-api: invalid getSingleContent")
	respondWithError(w, http.StatusNotFound, "Invalid ID")
}

func createContent(w http.ResponseWriter, r *http.Request) {
	var newContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		log.Print("helloworld-api: failed createContent")
	}
	json.Unmarshal(reqBody, &newContent)
	contents = append(contents, newContent)
	log.Print("helloworld-api: createContent received a request")
	respondWithJson(w, http.StatusOK, newContent)
}

func updateContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]
	var updatedContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print("helloworld-api: failed updateContent")
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	json.Unmarshal(reqBody, &updatedContent)
	for i, singleContent := range contents {
		if singleContent.ID == contentID {
			singleContent.Name = updatedContent.Name
			contents = append(contents[:i], singleContent)
			respondWithJson(w, http.StatusOK, singleContent)
		}
	}
	log.Print("helloworld-api: updateContent received a request")
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]

	for i, singleContent := range contents {
		if singleContent.ID == contentID {
			contents = append(contents[:i], contents[i+1:]...)
			log.Print("helloworld-api: deleteContent received a request")
			respondWithJson(w, http.StatusOK, "The content with has been deleted successfully")
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
