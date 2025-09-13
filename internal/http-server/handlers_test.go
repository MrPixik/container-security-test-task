package http_server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sandbox-dev-test-task/internal/models"
	"sandbox-dev-test-task/internal/pool"
	"testing"
)

func TestEnqueuePostHandler(t *testing.T) {
	p := pool.New(1)
	handler := enqueuePostHandler(p)

	reqBody := models.EnqueueReq{
		Id:         "1",
		Payload:    "hello",
		MaxRetries: 3,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var enqueRes models.EnqueueRes
	if err := json.NewDecoder(resp.Body).Decode(&enqueRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if enqueRes.TaskId == -1 {
		t.Errorf("expected valid value, got %d", enqueRes.TaskId)
	}
}

func TestEnqueuePostHandlerBadRequest(t *testing.T) {
	p := pool.New(1)
	handler := enqueuePostHandler(p)

	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader([]byte("{bad json}")))
	w := httptest.NewRecorder()

	handler(w, req)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for bad JSON, got %d", http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w := httptest.NewRecorder()

	healthCheckHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
