package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "hello", "code": "test"}

	respondJSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if result["message"] != "hello" || result["code"] != "test" {
		t.Errorf("unexpected response body: %v", result)
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()

	respondError(w, http.StatusBadRequest, "something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if result["error"] != "something went wrong" {
		t.Errorf("expected error message, got %q", result["error"])
	}
}
