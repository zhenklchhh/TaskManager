package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zhenklchhh/TaskManager/internal/domain"
)

const (
	SendWebhookTask = "send_webhook"
)

type WebhookPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

type WebhookTaskHandler struct {
	client *http.Client
}

func NewWebhookTaskHandler() *WebhookTaskHandler {
	return &WebhookTaskHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *WebhookTaskHandler) Handle(ctx context.Context, task *domain.Task) error {
	var payload WebhookPayload
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return fmt.Errorf("webhook handler: invalid json: %w", err)
	}

	if payload.URL == "" {
		return fmt.Errorf("webhook handler: URL is required")
	}

	if payload.Method == "" {
		payload.Method = http.MethodPost
	}

	if payload.Timeout > 0 {
		h.client.Timeout = time.Duration(payload.Timeout) * time.Second
	}

	var bodyReader io.Reader
	if payload.Body != nil {
		bodyBytes, err := json.Marshal(payload.Body)
		if err != nil {
			return fmt.Errorf("webhook handler: failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, bodyReader)
	if err != nil {
		return fmt.Errorf("webhook handler: failed to create request: %w", err)
	}

	if payload.Headers != nil {
		for key, value := range payload.Headers {
			req.Header.Set(key, value)
		}
	}

	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook handler: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook handler: unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
