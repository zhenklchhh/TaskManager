package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jhump/protoreflect/v2/grpcreflect"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

const GRPCTask = "grpc"

type GrpcTaskPayload struct {
	Address        string      `json:"address"`
	Service        string      `json:"service"`
	Method         string      `json:"method"`
	Message        interface{} `json:"message"`
	TimeoutSeconds int         `json:"timeout_seconds"`
	UseTls         bool        `json:"use_tls"`
}

type GrpcTaskHandler struct{}

func NewGrpcTaskHandler() *GrpcTaskHandler {
	return &GrpcTaskHandler{}
}

func (h *GrpcTaskHandler) Handle(ctx context.Context, t *domain.Task) error {
	var payload GrpcTaskPayload
	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		return fmt.Errorf("grpc: failed to unmarshal payload: %w", err)
	}

	if payload.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(payload.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(payload.Address, opts...)
	if err != nil {
		return fmt.Errorf("grpc: failed to dial %s: %w", payload.Address, err)
	}
	defer conn.Close()

	refClient := grpcreflect.NewClientAuto(ctx, conn)
	defer refClient.Reset()

	descriptor, err := refClient.AsResolver().FindDescriptorByName(protoreflect.FullName(payload.Service))
	if err != nil {
		return fmt.Errorf("grpc: failed to resolve service %s: %w", payload.Service, err)
	}

	svc, ok := descriptor.(protoreflect.ServiceDescriptor)
	if !ok {
		return fmt.Errorf("grpc: descriptor for %q is not a service", payload.Service)
	}

	method := svc.Methods().ByName(protoreflect.Name(payload.Method))
	if method == nil {
		return fmt.Errorf("grpc: method %s not found in service %s", payload.Method, payload.Service)
	}

	inputMsg := dynamicpb.NewMessage(method.Input())
	outputMsg := dynamicpb.NewMessage(method.Output())

	inputJSON, err := json.Marshal(payload.Message)
	if err != nil {
		return fmt.Errorf("grpc: failed to marshal message: %w", err)
	}

	if err := protojson.Unmarshal(inputJSON, inputMsg); err != nil {
		return fmt.Errorf("grpc: failed to unmarshal json to protobuf: %w", err)
	}

	fullMethod := fmt.Sprintf("/%s/%s", svc.FullName(), method.Name())
	slog.Info("grpc: invoking method", "method", fullMethod)

	if err := conn.Invoke(ctx, fullMethod, inputMsg, outputMsg); err != nil {
		return fmt.Errorf("grpc: invoke failed: %w", err)
	}

	return nil
}
