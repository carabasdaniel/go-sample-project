package app

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
	"github.com/aserto-dev/go-sample-project/pkg/app/impl"
	"github.com/aserto-dev/go-sample-project/pkg/app/server"
)

// GRPCServerRegistrations is where we register implementations with the GRPC server
func GRPCServerRegistrations(implInfo *impl.Info) server.Registrations {
	return func(server *grpc.Server) {
		info.RegisterInfoServer(server, implInfo)
	}
}

// GatewayServerRegistrations is where we register implementations with the Gateway server
func GatewayServerRegistrations() server.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {

		err := info.RegisterInfoHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			return errors.Wrap(err, "failed to register info handler with the gateway")
		}

		return nil
	}
}
