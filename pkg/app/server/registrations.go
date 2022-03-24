package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// Registrations represents a function that can register API implementations to the GRPC server.
type Registrations func(server *grpc.Server)

// HandlerRegistrations represents a function that can register handlers for the Gateway.
type HandlerRegistrations func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error
