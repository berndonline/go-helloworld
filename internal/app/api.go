package app

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"io/ioutil"
	"log"
	"net/http"
)

type api struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type allContent []api

func getIndexContent(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := StartSpanFromRequest("getIndexContent", tracer, r)
	childSpan := opentracing.GlobalTracer().StartSpan("http-response", opentracing.ChildOf(span.Context()))
	repo := getContentRepository()
	items, err := repo.ListContent(r.Context())
	if err != nil {
		log.Printf("helloworld: failed to list content from repository: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to list content")
		defer childSpan.Finish()
		defer span.Finish()
		return
	}
	log.Print("helloworld: getIndexContent received a request - " + getIPAddress(r))
	respondWithJson(w, http.StatusOK, items)
	defer childSpan.Finish()
	defer span.Finish()
}

func getSingleContent(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := StartSpanFromRequest("getSingleContent", tracer, r)
	childSpan := opentracing.GlobalTracer().StartSpan("http-response", opentracing.ChildOf(span.Context()))
	contentID := mux.Vars(r)["id"]
	repo := getContentRepository()
	content, err := repo.GetContent(r.Context(), contentID)
	if err != nil {
		defer childSpan.Finish()
		if errors.Is(err, ErrContentNotFound) {
			log.Print("helloworld: invalid getSingleContent")
			respondWithError(w, http.StatusNotFound, "Invalid ID")
			defer span.Finish()
			return
		}
		log.Printf("helloworld: failed getSingleContent: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch content")
		defer span.Finish()
		return
	}
	log.Print("helloworld: getSingleContent received a request - " + getIPAddress(r))
	respondWithJson(w, http.StatusOK, content)
	defer childSpan.Finish()
	defer span.Finish()
}

func createContent(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := StartSpanFromRequest("createContent", tracer, r)
	childSpan := opentracing.GlobalTracer().StartSpan("content-Insert", opentracing.ChildOf(span.Context()))
	defer r.Body.Close()
	var newContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		log.Print("helloworld: failed createContent")
		defer childSpan.Finish()
		defer span.Finish()
		return
	}
	json.Unmarshal(reqBody, &newContent)
	repo := getContentRepository()
	created, err := repo.CreateContent(r.Context(), newContent)
	if err != nil {
		if errors.Is(err, ErrContentAlreadyExists) {
			log.Print("helloworld: duplicate createContent")
			respondWithError(w, http.StatusConflict, "Content already exists")
			defer childSpan.Finish()
			defer span.Finish()
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		log.Print("helloworld: failed createContent")
		defer childSpan.Finish()
		defer span.Finish()
		return
	}
	log.Print("helloworld: createContent received a request - " + getIPAddress(r))
	respondWithJson(w, http.StatusCreated, created)
	defer childSpan.Finish()
	defer span.Finish()
}

func updateContent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	contentID := mux.Vars(r)["id"]
	var updatedContent api
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print("helloworld: failed updateContent")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.Unmarshal(reqBody, &updatedContent)
	repo := getContentRepository()
	updated, err := repo.UpdateContent(r.Context(), contentID, updatedContent.Name)
	if err != nil {
		if errors.Is(err, ErrContentNotFound) {
			respondWithError(w, http.StatusNotFound, "Invalid ID")
			return
		}
		log.Print("helloworld: failed updateContent")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, updated)
	log.Print("helloworld: updateContent received a request - " + getIPAddress(r))
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	contentID := mux.Vars(r)["id"]
	repo := getContentRepository()
	if err := repo.DeleteContent(r.Context(), contentID); err != nil {
		if errors.Is(err, ErrContentNotFound) {
			respondWithError(w, http.StatusNotFound, "Invalid ID")
			return
		}
		log.Printf("helloworld: failed deleteContent: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete content")
		return
	}
	log.Print("helloworld: deleteContent received a request - " + getIPAddress(r))
	respondWithJson(w, http.StatusOK, "The content with has been deleted successfully")
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
