package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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

	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getIndexContent", ext.RPCServerOption(spanCtx))

	if mongodb != false {
		childSpan := opentracing.GlobalTracer().StartSpan("mongodb", opentracing.ChildOf(span.Context()))
		contents, err := dao.FindAll()
    defer childSpan.Finish()
    subchildSpan := opentracing.GlobalTracer().StartSpan("http.response", opentracing.ChildOf(childSpan.Context()))

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: getIndexContent failed")
			defer subchildSpan.Finish()
			return
		}

		log.Print("helloworld-api: getIndexContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusOK, contents)
		defer subchildSpan.Finish()

	} else {

		childSpan := opentracing.GlobalTracer().StartSpan("http.response", opentracing.ChildOf(span.Context()))
		log.Print("helloworld-api: getIndexContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusOK, contents)
		defer childSpan.Finish()

	}
	defer span.Finish()
}

func getSingleContent(w http.ResponseWriter, r *http.Request) {

	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getSingleContent", ext.RPCServerOption(spanCtx))

	if mongodb != false {
		childSpan := opentracing.GlobalTracer().StartSpan("mongodb", opentracing.ChildOf(span.Context()))
		contentID, err := dao.FindById(mux.Vars(r)["id"])
		defer childSpan.Finish()
		subchildSpan := opentracing.GlobalTracer().StartSpan("http.response", opentracing.ChildOf(childSpan.Context()))

		if err != nil {
			respondWithError(w, http.StatusNotFound, "Invalid ID")
			log.Print("helloworld-api: getSingleContent invalid")
			defer subchildSpan.Finish()
			return
		}

		log.Print("helloworld-api: getSingleContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusOK, contentID)
		defer subchildSpan.Finish()

	} else {

		childSpan := opentracing.GlobalTracer().StartSpan("http.response", opentracing.ChildOf(span.Context()))
		contentID := mux.Vars(r)["id"]

		for _, singleContent := range contents {
			if singleContent.ID == contentID {
				log.Print("helloworld-api: getSingleContent received a request - " + getIPAddress(r))
				respondWithJson(w, http.StatusOK, singleContent)
				defer childSpan.Finish()
				defer span.Finish()
				return
			}
		}

		defer childSpan.Finish()
		log.Print("helloworld-api: invalid getSingleContent")
		respondWithError(w, http.StatusNotFound, "Invalid ID")
	}
	defer span.Finish()
}

func createContent(w http.ResponseWriter, r *http.Request) {

	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("createContent", ext.RPCServerOption(spanCtx))

	if mongodb != false {

    childSpan := opentracing.GlobalTracer().StartSpan("mongodb", opentracing.ChildOf(span.Context()))
		defer r.Body.Close()
		var newContent mgoApi

		if err := json.NewDecoder(r.Body).Decode(&newContent); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			log.Print("helloworld-api: createContent invalid request payload")
			defer childSpan.Finish()
			return
		}

		newContent.ID = bson.NewObjectId()
		if err := dao.Insert(newContent); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: createContent failed")
			defer childSpan.Finish()
			return
		}

		respondWithJson(w, http.StatusCreated, newContent)
		log.Print("helloworld-api: createContent received a request - " + getIPAddress(r))
		defer childSpan.Finish()

	} else {

    childSpan := opentracing.GlobalTracer().StartSpan("addedContent", opentracing.ChildOf(span.Context()))
		defer r.Body.Close()
		var newContent api
		reqBody, err := ioutil.ReadAll(r.Body)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			log.Print("helloworld-api: failed createContent")
			defer childSpan.Finish()
		}

		json.Unmarshal(reqBody, &newContent)
		contents = append(contents, newContent)
		log.Print("helloworld-api: createContent received a request - " + getIPAddress(r))
		respondWithJson(w, http.StatusCreated, newContent)
		defer childSpan.Finish()
	}
	defer span.Finish()
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
		respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
		log.Print("helloworld-api: deleteContent received a request - " + getIPAddress(r))

	} else {

		contentID := mux.Vars(r)["id"]
		for i, singleContent := range contents {
			if singleContent.ID == contentID {
				contents = append(contents[:i], contents[i+1:]...)
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
