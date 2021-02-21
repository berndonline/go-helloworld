package main

import (
	"bytes"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	// Below needed for APITEST library - https://apitest.dev
	// "github.com/steinfletcher/apitest"
	// jsonpath "github.com/steinfletcher/apitest-jsonpath"
)

// Test with standard go library

func Test_Standard_getContentIndex(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content", BasicAuth(getIndexContent, "Please enter your username and password")).Methods("GET")
	req, err := http.NewRequest("GET", "/api/v1/content", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(username, password)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `[{"id":"1","name":"Content 1"},{"id":"2","name":"Content 2"}]`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func Test_Standard_getSingleContent(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")

	t.Run("Find content 1", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/content/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"id":"1","name":"Content 1"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("Find content 2", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/content/2", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"id":"2","name":"Content 2"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/content/3", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"error":"Invalid ID"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})
}

func Test_Standard_CreateContent(t *testing.T) {
	r := mux.NewRouter()

	t.Run("Create content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content", BasicAuth(createContent, "Please enter your username and password")).Methods("POST")
		var jsonStr = []byte(`{"ID":"3","Name":"Content 3"}`)
		req, err := http.NewRequest("POST", "/api/v1/content", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
		expected := `{"id":"3","name":"Content 3"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("Find content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
		req, err := http.NewRequest("GET", "/api/v1/content/3", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"id":"3","name":"Content 3"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("Update content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content/{id}", BasicAuth(updateContent, "Please enter your username and password")).Methods("PUT")
		var jsonStr = []byte(`{"ID":"3","Name":"New content 3"}`)
		req, err := http.NewRequest("PUT", "/api/v1/content/3", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"id":"3","name":"New content 3"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})
	t.Run("Delete content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content/{id}", BasicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
		req, err := http.NewRequest("DELETE", "/api/v1/content/3", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		r.HandleFunc("/api/v1/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
		req, err := http.NewRequest("GET", "/api/v1/content/3", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := `{"error":"Invalid ID"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})
}

// Test using APITEST library - https://apitest.dev

// func Test_apitest_ContentIndex(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content", BasicAuth(getIndexContent, "Please enter your username and password")).Methods("GET")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    apitest.New().
//       // Report(apitest.SequenceDiagram("results")).
//       Handler(r).
//       Get("/api/v1/content").
//       BasicAuth("admin", "password").
//       Expect(t).
//       Status(http.StatusOK).
//       End()
// }
//
// func Test_apitest_SingleContent(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    t.Run("Find content 1", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Get("/api/v1/content/1").
//          BasicAuth("admin", "password").
//          Expect(t).
//          Assert(jsonpath.Equal(`$.Name`, "Content 1")).
//          Status(http.StatusOK).
//          End()
//    })
//    t.Run("Find content 2", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Get("/api/v1/content/2").
//          BasicAuth("admin", "password").
//          Expect(t).
//          Assert(jsonpath.Equal(`$.Name`, "Content 2")).
//          Status(http.StatusOK).
//          End()
//    })
//    t.Run("Not found", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Get("/api/v1/content/4").
//          BasicAuth("admin", "password").
//          Expect(t).
//          Status(http.StatusNotFound).
//          End()
//    })
// }
//
// func Test_apitest_createContent(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content", BasicAuth(createContent, "Please enter your username and password")).Methods("POST")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    t.Run("Create content 4", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Post("/api/v1/content").
//          JSON(`{"ID":"4","Name":"Content 4"}`).
//          BasicAuth("admin", "password").
//          Expect(t).
//          Status(http.StatusOK).
//          End()
//    })
// }
//
// func Test_apitest_checkNewContent(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    t.Run("Find content 4", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Get("/api/v1/content/4").
//          BasicAuth("admin", "password").
//          Expect(t).
//          Assert(jsonpath.Equal(`$.Name`, "Content 4")).
//          Status(http.StatusOK).
//          End()
//    })
// }
//
// func Test_apitest_updateContent(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content/{id}", BasicAuth(updateContent, "Please enter your username and password")).Methods("PUT")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    t.Run("Create content 4", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Put("/api/v1/content/4").
//          JSON(`{"ID":"4","Name":"New content 4"}`).
//          BasicAuth("admin", "password").
//          Expect(t).
//          Status(http.StatusOK).
//          End()
//    })
// }
//
// func Test_apitest_checkUpdateContent(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content/{id}", BasicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    t.Run("Find content 4", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Get("/api/v1/content/4").
//          BasicAuth("admin", "password").
//          Expect(t).
//          Assert(jsonpath.Equal(`$.Name`, "New content 4")).
//          Status(http.StatusOK).
//          End()
//    })
// }
//
// func Test_apitest_deleteContent(t *testing.T) {
//    r := mux.NewRouter()
//    r.HandleFunc("/api/v1/content/{id}", BasicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
//    ts := httptest.NewServer(r)
//    defer ts.Close()
//    t.Run("Delete content 4", func(t *testing.T) {
//       apitest.New().
//          // Report(apitest.SequenceDiagram("results")).
//          Handler(r).
//          Delete("/api/v1/content/4").
//          BasicAuth("admin", "password").
//          Expect(t).
//          Status(http.StatusOK).
//          End()
//    })
// }
