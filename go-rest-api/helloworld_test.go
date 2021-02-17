package main

import (
   "net/http"
   "net/http/httptest"
   "testing"
   "github.com/gorilla/mux"
   "github.com/steinfletcher/apitest"
   jsonpath "github.com/steinfletcher/apitest-jsonpath"
)

// Test with standard go library

func Test_Standard_getDefaultHandler(t *testing.T) {
  r := mux.NewRouter()
  r.HandleFunc("/", handler)
  ts := httptest.NewServer(r)
  defer ts.Close()
  res, err := http.Get(ts.URL + "")
  if err != nil {
     t.Errorf("Expected nil, received %s", err.Error())
  }
  if res.StatusCode != http.StatusOK {
     t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
  }
}

func Test_Standard_getContentIndex(t *testing.T) {
  r := mux.NewRouter()
  r.HandleFunc("/api/v1/content", BasicAuth(getAllContent, "Please enter your username and password")).Methods("GET")
  ts := httptest.NewServer(r)
  defer ts.Close()
  client := ts.Client()
  req, err := http.NewRequest("GET", ts.URL + "/api/v1/content", nil)
  if err != nil {
    t.Fatal(err)
  }
  req.SetBasicAuth(ADMIN_USER, ADMIN_PASSWORD)
  res, err := client.Do(req)
  if err != nil {
     t.Errorf("Expected nil, received %s", err.Error())
  }
  if res.StatusCode != http.StatusOK {
     t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
  }
}

func Test_Standard_getOneContent(t *testing.T) {
  r := mux.NewRouter()
  r.HandleFunc("/api/v1/content/{id}", BasicAuth(getOneContent, "Please enter your username and password")).Methods("GET")
  ts := httptest.NewServer(r)
  defer ts.Close()
  client := ts.Client()
  t.Run("Find content 1", func(t *testing.T) {
    req, err := http.NewRequest("GET", ts.URL + "/api/v1/content/1", nil)
    if err != nil {
      t.Fatal(err)
    }
    req.SetBasicAuth(ADMIN_USER, ADMIN_PASSWORD)
    res, err := client.Do(req)
         if err != nil {
            t.Errorf("Expected nil, received %s", err.Error())
         }
         if res.StatusCode != http.StatusOK {
            t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
         }
  })
  t.Run("Find content 2", func(t *testing.T) {
    req, err := http.NewRequest("GET", ts.URL + "/api/v1/content/2", nil)
    if err != nil {
      t.Fatal(err)
    }
    req.SetBasicAuth(ADMIN_USER, ADMIN_PASSWORD)
    res, err := client.Do(req)
         if err != nil {
            t.Errorf("Expected nil, received %s", err.Error())
         }
         if res.StatusCode != http.StatusOK {
            t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
         }
  })
  t.Run("Not Found", func(t *testing.T) {
    req, err := http.NewRequest("GET", ts.URL + "/api/v1/content/3", nil)
    if err != nil {
      t.Fatal(err)
    }
    req.SetBasicAuth(ADMIN_USER, ADMIN_PASSWORD)
     res, err := client.Do(req)
     if err != nil {
        t.Errorf("Expected nil, received %s", err.Error())
     }
     if res.StatusCode != http.StatusNotFound {
        t.Errorf("Expected %d, received %d", http.StatusNotFound, res.StatusCode)
     }
  })
}

// Test using APITEST library - https://apitest.dev

func Test_apitest_getDefaultHandler(t *testing.T) {
  r := mux.NewRouter()
  r.HandleFunc("/", handler)
  ts := httptest.NewServer(r)
  defer ts.Close()
  apitest.New().
     // Report(apitest.SequenceDiagram("results")).
     Handler(r).
     Get("/").
     Expect(t).
     Status(http.StatusOK).
     End()
}

func Test_apitest_getContentIndex(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content", BasicAuth(getAllContent, "Please enter your username and password")).Methods("GET")
   ts := httptest.NewServer(r)
   defer ts.Close()
   apitest.New().
      // Report(apitest.SequenceDiagram("results")).
      Handler(r).
      Get("/api/v1/content").
      BasicAuth("admin", "password").
      Expect(t).
      Status(http.StatusOK).
      End()
}

func Test_apitest_getOneContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content/{id}", BasicAuth(getOneContent, "Please enter your username and password")).Methods("GET")
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("Find content 1", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Get("/api/v1/content/1").
         BasicAuth("admin", "password").
         Expect(t).
         Assert(jsonpath.Equal(`$.Name`, "Content 1")).
         Status(http.StatusOK).
         End()
   })
   t.Run("Find content 2", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Get("/api/v1/content/2").
         BasicAuth("admin", "password").
         Expect(t).
         Assert(jsonpath.Equal(`$.Name`, "Content 2")).
         Status(http.StatusOK).
         End()
   })
   t.Run("Not found", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Get("/api/v1/content/3").
         BasicAuth("admin", "password").
         Expect(t).
         Status(http.StatusNotFound).
         End()
   })
}

func Test_apitest_createContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content", BasicAuth(createContent, "Please enter your username and password")).Methods("POST")
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("Create content 3", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Post("/api/v1/content").
         JSON(`{"ID":"3","Name":"Content 3"}`).
         BasicAuth("admin", "password").
         Expect(t).
         Status(http.StatusOK).
         End()
   })
}

func Test_apitest_checkNewContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content/{id}", BasicAuth(getOneContent, "Please enter your username and password")).Methods("GET")
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("Find content 3", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Get("/api/v1/content/3").
         BasicAuth("admin", "password").
         Expect(t).
         Assert(jsonpath.Equal(`$.Name`, "Content 3")).
         Status(http.StatusOK).
         End()
   })
}

func Test_apitest_updateContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content/{id}", BasicAuth(updateContent, "Please enter your username and password")).Methods("PUT")
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("Create content 3", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Put("/api/v1/content/3").
         JSON(`{"ID":"3","Name":"New content 3"}`).
         BasicAuth("admin", "password").
         Expect(t).
         Status(http.StatusOK).
         End()
   })
}

func Test_apitest_checkUpdateContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content/{id}", BasicAuth(getOneContent, "Please enter your username and password")).Methods("GET")
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("Find content 3", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Get("/api/v1/content/3").
         BasicAuth("admin", "password").
         Expect(t).
         Assert(jsonpath.Equal(`$.Name`, "New content 3")).
         Status(http.StatusOK).
         End()
   })
}

func Test_apitest_deleteContent(t *testing.T) {
   r := mux.NewRouter()
   r.HandleFunc("/api/v1/content/{id}", BasicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
   ts := httptest.NewServer(r)
   defer ts.Close()
   t.Run("Create content 3", func(t *testing.T) {
      apitest.New().
         // Report(apitest.SequenceDiagram("results")).
         Handler(r).
         Delete("/api/v1/content/3").
         BasicAuth("admin", "password").
         Expect(t).
         Status(http.StatusOK).
         End()
   })
}
