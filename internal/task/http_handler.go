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

const HttpTask = "http"

type HttpTaskPayload struct {
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Body            interface{}       `json:"body"`
	Headers         map[string]string `json:"headers"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	RetryOnStatuses []int             `json:"retry_on"`
}

type HttpHandler struct {
	client http.Client
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{
		client: http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *HttpHandler) Handle(ctx context.Context, t *domain.Task) error {
	var payload HttpTaskPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return httpHandlerError(err)
	}
	if payload.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(payload.TimeoutSeconds) * time.Second)
		defer cancel()
	}
	var body io.Reader
	if payload.Body != nil {
		bodyBytes, err := json.Marshal(payload.Body)
		if err != nil {
			return httpHandlerError(err)
		}
		body = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, body)
	if err != nil {
		return httpHandlerError(err)
	}
	for k, v := range payload.Headers {
		req.Header.Set(k, v)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return httpHandlerError(err)
	}
	defer resp.Body.Close()
	for _, status := range payload.RetryOnStatuses {
		if resp.StatusCode == status {
			return fmt.Errorf("http error: %d", resp.StatusCode)
		}
	}
	return nil
}

func httpHandlerError(err error) error {
	return fmt.Errorf("http handler error: ", err)
}
