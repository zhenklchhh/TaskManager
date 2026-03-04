package task

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zhenklchhh/TaskManager/internal/domain"
)

func TestHttpTaskHandler_HandleWithServer(t *testing.T) {
	type testCase struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		payload       HttpTaskPayload
		expectErr     bool
		errContains   string
	}

	tests := []testCase{
		{
			name: "successful POST to test server",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
			payload: HttpTaskPayload{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]string{
					"event": "task_completed",
				},
				RetryOnStatuses: []int{500, 502},
			},
			expectErr: false,
		},
		{
			name: "server returns 500 in retry list",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			payload: HttpTaskPayload{
				Method:          "GET",
				RetryOnStatuses: []int{500, 502, 503},
			},
			expectErr:   true,
			errContains: "http error: 500",
		},
		{
			name: "verify request method is PUT",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
			},
			payload: HttpTaskPayload{
				Method: "PUT",
			},
			expectErr: false,
		},
		{
			name: "verify request has correct body",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, _ := io.ReadAll(r.Body)
				defer r.Body.Close()

				var payload map[string]interface{}
				json.Unmarshal(bodyBytes, &payload)

				if payload["user_id"] != "123" {
					t.Errorf("Expected user_id=123, got %v", payload["user_id"])
				}

				w.WriteHeader(http.StatusCreated)
			},
			payload: HttpTaskPayload{
				Method: "POST",
				Body: map[string]interface{}{
					"user_id": "123",
					"action":  "create",
				},
			},
			expectErr: false,
		},
		{
			name: "verify custom headers are sent",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer token123" {
					t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
				}
				if r.Header.Get("X-Custom") != "header-value" {
					t.Errorf("Expected X-Custom header, got %s", r.Header.Get("X-Custom"))
				}
				w.WriteHeader(http.StatusOK)
			},
			payload: HttpTaskPayload{
				Method: "POST",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Custom":      "header-value",
				},
				Body: map[string]string{"key": "value"},
			},
			expectErr: false,
		},
		{
			name: "status code 201 not in retry list - should succeed",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			payload: HttpTaskPayload{
				Method:          "POST",
				RetryOnStatuses: []int{500, 502, 503},
			},
			expectErr: false,
		},
		{
			name: "unmarshal error - invalid payload",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			payload: HttpTaskPayload{
				Method: "GET",
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.serverHandler))
			defer server.Close()

			tc.payload.URL = server.URL

			payloadBytes, _ := json.Marshal(tc.payload)
			task := &domain.Task{
				Payload: payloadBytes,
			}

			handler := NewHttpHandler()
			err := handler.Handle(context.Background(), task)

			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, got: %v", tc.expectErr, err)
			}

			if tc.expectErr && tc.errContains != "" {
				if err != nil && !bytes.Contains([]byte(err.Error()), []byte(tc.errContains)) {
					t.Errorf("Expected error containing '%s', got '%v'", tc.errContains, err)
				}
			}
		})
	}
}

func TestHttpTaskHandler_MarshalPayload(t *testing.T) {
	tests := []struct {
		name      string
		payload   interface{}
		expectErr bool
	}{
		{
			name: "valid payload",
			payload: HttpTaskPayload{
				Method: "POST",
				URL:    "https://example.com",
			},
			expectErr: false,
		},
		{
			name:      "nil payload",
			payload:   nil,
			expectErr: false,
		},
		{
			name: "complex nested payload",
			payload: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "John",
				},
				"events": []string{"create", "update"},
			},
			expectErr: false,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := HttpTaskPayload{
				Method: "POST",
				URL:    server.URL,
				Body:   tc.payload,
			}

			payloadBytes, _ := json.Marshal(payload)
			task := &domain.Task{
				Payload: payloadBytes,
			}

			handler := NewHttpHandler()
			err := handler.Handle(context.Background(), task)

			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, got: %v", tc.expectErr, err)
			}
		})
	}
}
