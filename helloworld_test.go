package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	// Below needed for APITEST library - https://apitest.dev
	// "github.com/steinfletcher/apitest"
)

// Test with standard go library

func Test_Standard_Handler(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `Hello, World - REST API!` + "\n" + os.Getenv("HOSTNAME")
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func Test_Standard_Healthz(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz)
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

// Test using APITEST library - https://apitest.dev

// func Test_apitest_Handler(t *testing.T) {
//   r := mux.NewRouter()
//   r.HandleFunc("/", handler)
//   ts := httptest.NewServer(r)
//   defer ts.Close()
//   apitest.New().
//      // Report(apitest.SequenceDiagram("results")).
//      Handler(r).
//      Get("/").
//      Expect(t).
//      Status(http.StatusOK).
//      End()
// }
