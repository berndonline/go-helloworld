package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
)

type api struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mgoApi struct {
	ID   bson.ObjectId `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
}

type allContent []api

var contents = allContent{
	{
		ID:   "1",
		Name: "Content 1",
	},
	{
		ID:   "2",
		Name: "Content 2",
	},
}

func getIndexContent(w http.ResponseWriter, r *http.Request) {

	if mongodb != false {
		contents, err := dao.FindAll()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: getIndexContent failed")
			return
		}
		go TracingExtract(r, "getIndexContent-mgoApi")
		log.Print("helloworld-api: getIndexContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusOK, contents)

	} else {
    go TracingExtract(r, "getIndexContent")
		log.Print("helloworld-api: getIndexContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusOK, contents)
	}
}

func getSingleContent(w http.ResponseWriter, r *http.Request) {
	if mongodb != false {
		contentID, err := dao.FindById(mux.Vars(r)["id"])
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Invalid ID")
			log.Print("helloworld-api: getSingleContent invalid")
			return
		}
		go TracingExtract(r, "getSingleContent-mgoApi")
		log.Print("helloworld-api: getSingleContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusOK, contentID)

	} else {
		contentID := mux.Vars(r)["id"]
		for _, singleContent := range contents {
			if singleContent.ID == contentID {
				go TracingExtract(r, "getSingleContent")
				log.Print("helloworld-api: getSingleContent received a request - " + getIPAddress(r))
				respondWithJson(w, http.StatusOK, singleContent)
				return
			}
		}
		log.Print("helloworld-api: invalid getSingleContent")
		respondWithError(w, http.StatusNotFound, "Invalid ID")
	}
}

func createContent(w http.ResponseWriter, r *http.Request) {
	if mongodb != false {

		defer r.Body.Close()
		var newContent mgoApi
		if err := json.NewDecoder(r.Body).Decode(&newContent); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			log.Print("helloworld-api: createContent invalid request payload")
			return
		}
		newContent.ID = bson.NewObjectId()
		if err := dao.Insert(newContent); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: createContent failed")
			return
		}
		go TracingExtract(r, "createContent-mgoApi")
		respondWithJson(w, http.StatusCreated, newContent)
		log.Print("helloworld-api: createContent received a request - " + getIPAddress(r))

	} else {

		defer r.Body.Close()
		var newContent api
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: failed createContent")
		}
		json.Unmarshal(reqBody, &newContent)
		contents = append(contents, newContent)
		go TracingExtract(r, "createContent")
		log.Print("helloworld-api: createContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusCreated, newContent)
	}
}

func updateContent(w http.ResponseWriter, r *http.Request) {
	if mongodb != false {

		defer r.Body.Close()
		var updatedContent mgoApi
		updatedContent.ID = bson.ObjectIdHex(mux.Vars(r)["id"])
		if err := json.NewDecoder(r.Body).Decode(&updatedContent); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			log.Print("helloworld-api: updateContent invalid request payload")
			return
		}
		if err := dao.Update(updatedContent); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: updateContent failed")
			return
		}
		go TracingExtract(r, "updateContent-mgoApi")
		respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
		log.Print("helloworld-api: updateContent received a request - " + getIPAddress(r))

	} else {

		defer r.Body.Close()
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
		go TracingExtract(r, "updateContent")
		log.Print("helloworld-api: updateContent received a request - " + getIPAddress(r))
	}
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
	if mongodb != false {

		content, err := dao.FindById(mux.Vars(r)["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid ID")
			log.Print("helloworld-api: deleteContent invalid id")
			return
		}
		if err := dao.Delete(content); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: deleteContent failed")
			return
		}
		go TracingExtract(r, "deleteContent-mgoApi")
		respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
		log.Print("helloworld-api: deleteContent received a request - " + getIPAddress(r))

	} else {

		contentID := mux.Vars(r)["id"]
		for i, singleContent := range contents {
			if singleContent.ID == contentID {
				contents = append(contents[:i], contents[i+1:]...)
				go TracingExtract(r, "deleteContent")
				log.Print("helloworld-api: deleteContent received a request - " + getIPAddress(r))
				respondWithJson(w, http.StatusOK, "The content with has been deleted successfully")
			}
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
