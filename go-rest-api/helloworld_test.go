package main

import (
   "net/http"
   "net/http/httptest"
   "testing"

   "github.com/gorilla/mux"
)

func Test_ContentIndex(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content", getAllContent)
   ts := httptest.NewServer(r)
   defer ts.Close()
   res, err := http.Get(ts.URL + "/api/v1/content")
   if err != nil {
      t.Errorf("Expected nil, received %s", err.Error())
   }
   if res.StatusCode != http.StatusOK {
      t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
   }
}

func Test_GetOneContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content/{id}", getOneContent)
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("not found", func(t *testing.T) {
      res, err := http.Get(ts.URL + "/api/v1/content/3")
      if err != nil {
         t.Errorf("Expected nil, received %s", err.Error())
      }
      if res.StatusCode != http.StatusNotFound {
         t.Errorf("Expected %d, received %d", http.StatusNotFound, res.StatusCode)
      }
   })
   t.Run("found", func(t *testing.T) {
      res, err := http.Get(ts.URL + "/api/v1/content/1")
      if err != nil {
         t.Errorf("Expected nil, received %s", err.Error())
      }
      if res.StatusCode != http.StatusOK {
         t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
      }
   })
}
