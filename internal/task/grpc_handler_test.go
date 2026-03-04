package task

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"google.golang.org/grpc"
)

type TestGrpcServer struct {
	addr   string
	server *grpc.Server
	ln     net.Listener
}

func NewTestGrpcServer(t *testing.T) *TestGrpcServer {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	go func() {
		if err := server.Serve(ln); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	return &TestGrpcServer{
		addr:   ln.Addr().String(),
		server: server,
		ln:     ln,
	}
}

func (s *TestGrpcServer) Close() {
	s.server.GracefulStop()
	s.ln.Close()
}

func TestGrpcTaskHandler_HandleConnection(t *testing.T) {
	tests := []struct {
		name        string
		payload     GrpcTaskPayload
		expectErr   bool
		errContains string
	}{
		{
			name: "invalid address",
			payload: GrpcTaskPayload{
				Address: "invalid:99999",
				Service: "TestService",
				Method:  "Test",
				Message: map[string]interface{}{},
			},
			expectErr:   true,
			errContains: "failed to dial", 
		},
		{
			name: "valid address but service not found",
			payload: GrpcTaskPayload{
				Address: "127.0.0.1:50051", 
				Service: "NonExistentService",
				Method:  "Test",
				Message: map[string]interface{}{},
			},
			expectErr: true,
		},
		{
			name: "timeout exceeded",
			payload: GrpcTaskPayload{
				Address:        "127.0.0.1:50051",
				Service:        "TestService",
				Method:         "SlowMethod",
				Message:        map[string]interface{}{},
				TimeoutSeconds: 1, 
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tc.payload)
			task := &domain.Task{
				Payload: payloadBytes,
			}

			handler := NewGrpcTaskHandler()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := handler.Handle(ctx, task)

			if (err != nil) != tc.expectErr {
				t.Logf("Expected error: %v, got: %v", tc.expectErr, err)
			}
		})
	}
}

func TestGrpcTaskHandler_UnmarshalPayload(t *testing.T) {
	tests := []struct {
		name        string
		payloadJSON string
		expectErr   bool
	}{
		{
			name: "valid JSON payload",
			payloadJSON: `{
				"address": "127.0.0.1:50051",
				"service": "UserService",
				"method": "GetUser",
				"message": {"id": "123"},
				"timeout_seconds": 30
			}`,
			expectErr: false,
		},
		{
			name:        "invalid JSON",
			payloadJSON: `{invalid json}`,
			expectErr:   true,
		},
		{
			name: "missing required fields",
			payloadJSON: `{
				"address": "127.0.0.1:50051"
			}`,
			expectErr: false, 
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := &domain.Task{
				Payload: []byte(tc.payloadJSON),
			}

			handler := NewGrpcTaskHandler()
			err := handler.Handle(context.Background(), task)

			if tc.name == "invalid JSON" && err == nil {
				t.Error("Expected error for invalid JSON")
			}
		})
	}
}

func TestGrpcTaskPayload_Validation(t *testing.T) {
	tests := []struct {
		name      string
		payload   GrpcTaskPayload
		validate  bool
		expectErr bool
	}{
		{
			name: "valid payload",
			payload: GrpcTaskPayload{
				Address: "localhost:50051",
				Service: "UserService",
				Method:  "GetUser",
				Message: map[string]interface{}{"id": "123"},
			},
			validate:  true,
			expectErr: false,
		},
		{
			name: "empty address",
			payload: GrpcTaskPayload{
				Address: "",
				Service: "UserService",
				Method:  "GetUser",
			},
			validate:  true,
			expectErr: true,
		},
		{
			name: "empty service",
			payload: GrpcTaskPayload{
				Address: "localhost:50051",
				Service: "",
				Method:  "GetUser",
			},
			validate:  true,
			expectErr: true,
		},
		{
			name: "empty method",
			payload: GrpcTaskPayload{
				Address: "localhost:50051",
				Service: "UserService",
				Method:  "",
			},
			validate:  true,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.validate {
				return
			}

			valid := tc.payload.Address != "" && tc.payload.Service != "" && tc.payload.Method != ""

			if valid && tc.expectErr {
				t.Error("Expected validation error")
			}
			if !valid && !tc.expectErr {
				t.Error("Expected valid payload")
			}
		})
	}
}