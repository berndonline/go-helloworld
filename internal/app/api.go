package app

import (
    "encoding/json"
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
    span := StartSpanFromRequest("getIndexContent", tracer, r)
    childSpan := opentracing.GlobalTracer().StartSpan("http-response", opentracing.ChildOf(span.Context()))
    log.Print("helloworld: getIndexContent received a request - " + getIPAddress(r))
    respondWithJson(w, http.StatusOK, contents)
    defer childSpan.Finish()
    defer span.Finish()
}

func getSingleContent(w http.ResponseWriter, r *http.Request) {
    tracer := opentracing.GlobalTracer()
    span := StartSpanFromRequest("getSingleContent", tracer, r)
    childSpan := opentracing.GlobalTracer().StartSpan("http-response", opentracing.ChildOf(span.Context()))
    contentID := mux.Vars(r)["id"]
    for _, singleContent := range contents {
        if singleContent.ID == contentID {
            log.Print("helloworld: getSingleContent received a request - " + getIPAddress(r))
            respondWithJson(w, http.StatusOK, singleContent)
            defer childSpan.Finish()
            defer span.Finish()
            return
        }
    }
    defer childSpan.Finish()
    log.Print("helloworld: invalid getSingleContent")
    respondWithError(w, http.StatusNotFound, "Invalid ID")
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
    contents = append(contents, newContent)
    log.Print("helloworld: createContent received a request - " + getIPAddress(r))
    respondWithJson(w, http.StatusCreated, newContent)
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
    for i, singleContent := range contents {
        if singleContent.ID == contentID {
            singleContent.Name = updatedContent.Name
            contents = append(contents[:i], singleContent)
            respondWithJson(w, http.StatusOK, singleContent)
            log.Print("helloworld: updateContent received a request - " + getIPAddress(r))
            return
        }
    }
    respondWithError(w, http.StatusNotFound, "Invalid ID")
}

func deleteContent(w http.ResponseWriter, r *http.Request) {
    contentID := mux.Vars(r)["id"]
    for i, singleContent := range contents {
        if singleContent.ID == contentID {
            contents = append(contents[:i], contents[i+1:]...)
            log.Print("helloworld: deleteContent received a request - " + getIPAddress(r))
            respondWithJson(w, http.StatusOK, "The content with has been deleted successfully")
            return
        }
    }
    respondWithError(w, http.StatusNotFound, "Invalid ID")
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
