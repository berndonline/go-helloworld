package main

import (
	"bytes"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	username = "user1"
	password = "password1"
)

func Test_getContentIndex(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content", basicAuth(getIndexContent, "Please enter your username and password")).Methods("GET")
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

func Test_getSingleContent(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content/{id}", basicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")

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

func Test_CreateContent(t *testing.T) {
	r := mux.NewRouter()

	t.Run("Create content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content", basicAuth(createContent, "Please enter your username and password")).Methods("POST")
		var jsonStr = []byte(`{"id":"3","name":"Content 3"}`)
		req, err := http.NewRequest("POST", "/api/v1/content", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(username, password)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusCreated)
		}
		expected := `{"id":"3","name":"Content 3"}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("Find content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content/{id}", basicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(updateContent, "Please enter your username and password")).Methods("PUT")
		var jsonStr = []byte(`{"id":"3","name":"New content 3"}`)
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(deleteContent, "Please enter your username and password")).Methods("DELETE")
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(getSingleContent, "Please enter your username and password")).Methods("GET")
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

func Test_jwtChecks(t *testing.T) {
	r := mux.NewRouter()

	t.Run("JWT Login", func(t *testing.T) {
		r.HandleFunc("/api/v2/login", jwtLogin).Methods("POST")
		var jsonStr = []byte(`{"username":"user1","password":"password1"}`)
		req, err := http.NewRequest("POST", "/api/v2/login", bytes.NewBuffer(jsonStr))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := "Token issued.\n"
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})
}
