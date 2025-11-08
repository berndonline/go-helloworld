package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/mux"
)

var (
	username = "user1"
	password = "password1"
)

func resetRepository() ContentRepository {
	repo := newInMemoryRepository(nil)
	setContentRepository(repo)
	return repo
}

type mockPublisher struct {
	mu     sync.Mutex
	events []api
}

func (m *mockPublisher) Publish(_ context.Context, item api) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, item)
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func (m *mockPublisher) published() []api {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]api, len(m.events))
	copy(cp, m.events)
	return cp
}

func Test_getContentIndex(t *testing.T) {
	repo := resetRepository()
	ctx := context.Background()
	if _, err := repo.CreateContent(ctx, api{ID: "1", Name: "Content 1"}); err != nil {
		t.Fatalf("failed to seed content 1: %v", err)
	}
	if _, err := repo.CreateContent(ctx, api{ID: "2", Name: "Content 2"}); err != nil {
		t.Fatalf("failed to seed content 2: %v", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content", basicAuth(getIndexContent)).Methods("GET")
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
	repo := resetRepository()
	ctx := context.Background()
	if _, err := repo.CreateContent(ctx, api{ID: "1", Name: "Content 1"}); err != nil {
		t.Fatalf("failed to seed content 1: %v", err)
	}
	if _, err := repo.CreateContent(ctx, api{ID: "2", Name: "Content 2"}); err != nil {
		t.Fatalf("failed to seed content 2: %v", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content/{id}", basicAuth(getSingleContent)).Methods("GET")

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
	resetRepository()
	r := mux.NewRouter()

	t.Run("Create content 3", func(t *testing.T) {
		r.HandleFunc("/api/v1/content", basicAuth(createContent)).Methods("POST")
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(getSingleContent)).Methods("GET")
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(updateContent)).Methods("PUT")
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(deleteContent)).Methods("DELETE")
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
		r.HandleFunc("/api/v1/content/{id}", basicAuth(getSingleContent)).Methods("GET")
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

func Test_CreateContentPublishesEvent(t *testing.T) {
	resetRepository()
	pub := &mockPublisher{}
	setContentPublisher(pub)
	defer resetContentPublisher()

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/content", basicAuth(createContent)).Methods("POST")

	reqBody := []byte(`{"id":"42","name":"Kafka Hello"}`)
	req, err := http.NewRequest("POST", "/api/v1/content", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusCreated)
	}
	events := pub.published()
	if len(events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(events))
	}
	if events[0].ID != "42" || events[0].Name != "Kafka Hello" {
		t.Fatalf("unexpected payload published: %+v", events[0])
	}
}

func Test_jwtChecks(t *testing.T) {
	resetRepository()
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
