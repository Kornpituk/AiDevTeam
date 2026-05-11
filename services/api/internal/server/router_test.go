package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCORSPreflightReturnsOK(t *testing.T) {
	for _, origin := range []string{"http://localhost:3000", "http://localhost:3001"} {
		t.Run(origin, func(t *testing.T) {
			router := mux.NewRouter()
			router.Use(corsMiddleware)
			router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			router.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			}).Methods(http.MethodPost)

			req := httptest.NewRequest(http.MethodOptions, "/tasks", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			req.Header.Set("Access-Control-Request-Headers", "Content-Type")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", origin, got)
			}

			if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
				t.Fatal("expected Access-Control-Allow-Methods header")
			}
		})
	}
}
