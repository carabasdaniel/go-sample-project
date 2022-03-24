package server

import (
	"time"

	"github.com/aserto-dev/go-utils/certs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
)

// newGRPCServer sets up a new GRPC server
func newGRPCServer(cfg *config.Config, logger *zerolog.Logger, registrations Registrations) (*grpc.Server, error) {
	grpc.EnableTracing = true

	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		logger.Error().Err(err).Msg("failed to register ocgrpc server views")
	}

	connectionTimeout := time.Duration(cfg.API.GRPC.ConnectionTimeoutSeconds) * time.Second
	tlsCreds, err := certs.GRPCServerTLSCreds(cfg.API.GRPC.Certs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate tls config")
	}

	tlsAuth := grpc.Creds(tlsCreds)
	server := grpc.NewServer(
		tlsAuth,
		grpc.ConnectionTimeout(connectionTimeout),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)
	reflection.Register(server)

	registrations(server)

	return server, nil
}
