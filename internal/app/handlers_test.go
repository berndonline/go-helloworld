package app

import (
    "github.com/gorilla/mux"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
)

func Test_Handlers(t *testing.T) {
    r := mux.NewRouter()
    r.HandleFunc("/", handler)
    r.HandleFunc("/healthz", healthz)
    r.HandleFunc("/readyz", readyz)

    t.Run("Test default handler", func(t *testing.T) {
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
    })
    t.Run("Test Healthz handler", func(t *testing.T) {
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

        expected := `ok`
        if rr.Body.String() != expected {
            t.Errorf("handler returned unexpected body: got %v want %v",
                rr.Body.String(), expected)
        }
    })
    t.Run("Test Readyz handler", func(t *testing.T) {
        req, err := http.NewRequest("GET", "/readyz", nil)
        if err != nil {
            t.Fatal(err)
        }

        rr := httptest.NewRecorder()
        r.ServeHTTP(rr, req)

        if status := rr.Code; status != http.StatusOK {
            t.Errorf("handler returned wrong status code: got %v want %v",
                status, http.StatusOK)
        }

        expected := `ok`
        if rr.Body.String() != expected {
            t.Errorf("handler returned unexpected body: got %v want %v",
                rr.Body.String(), expected)
        }
    })
}

