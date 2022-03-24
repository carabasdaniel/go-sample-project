package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/aserto-dev/go-utils/certs"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/aserto-dev/go-sample-project/pkg/cc"
)

const (
	svcName = "go-sample-project"
)

// Server manages the GRPC and HTTP servers, as well as their health servers.
type Server struct {
	*cc.CC
	logger       *zerolog.Logger
	grpcServer   *grpc.Server
	gtwServer    *http.Server
	healthServer *HealthServer
	gtwMux       *runtime.ServeMux

	handlerRegistrations HandlerRegistrations
}

// NewServer sets up a new server
func NewServer(
	c *cc.CC,
	registrations Registrations,
	handlerRegistrations HandlerRegistrations,
) (*Server, func(), error) {
	newLogger := c.Log.With().Str("component", fmt.Sprintf("api.%s", svcName)).Logger()

	grpcServer, err := newGRPCServer(c.Config, &newLogger, registrations)
	if err != nil {
		return nil, nil, err
	}

	gtwMux := gatewayMux()
	gtwServer, err := newGatewayServer(&newLogger, c.Config, gtwMux, c.MetricsRecorder)
	if err != nil {
		return nil, nil, err
	}

	healthServer := newGRPCHealthServer()

	server := &Server{
		CC:                   c,
		logger:               &newLogger,
		grpcServer:           grpcServer,
		gtwServer:            gtwServer,
		gtwMux:               gtwMux,
		healthServer:         healthServer,
		handlerRegistrations: handlerRegistrations,
	}

	return server, func() {
		err := server.Stop()
		if err != nil {
			newLogger.Error().Err(err).Msg("failed to stop server")
		}
	}, nil
}

// Start starts the GRPC and HTTP servers, as well as their health servers.
func (s *Server) Start() error {
	s.logger.Info().Msg("server::Start")

	grpc.EnableTracing = true

	if err := s.startHealthService(s.Config.API.Health.ListenAddress); err != nil {
		return errors.Wrap(err, "failed to start health server")
	}

	if err := s.startGRPCServer(s.Config.API.GRPC.ListenAddress); err != nil {
		return errors.Wrap(err, "failed to start grpc server")
	}

	if err := s.startGatewayServer(s.Config.API.Gateway.ListenAddress); err != nil {
		return errors.Wrap(err, "failed to start gateway server")
	}

	s.healthServer.Server.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", svcName), healthpb.HealthCheckResponse_SERVING)

	return nil
}

// Stop stops the GRPC and HTTP servers, as well as their health servers.
func (s *Server) Stop() error {
	var result error

	s.logger.Info().Msg("Server stopping.")

	if s.gtwServer != nil {
		err := s.stopHTTPServer(s.gtwServer)
		if err != nil {
			result = multierror.Append(result, errors.Wrap(err, "failed to stop gateway server"))
		}
	}

	if s.healthServer != nil {
		s.healthServer.Server.SetServingStatus(
			fmt.Sprintf("grpc.health.v1.%s", svcName),
			healthpb.HealthCheckResponse_NOT_SERVING,
		)
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.healthServer.GRPCServer != nil {
		s.healthServer.GRPCServer.GracefulStop()
	}

	err := s.ErrGroup.Wait()
	if err != nil {
		s.logger.Info().Err(err).Msg("shutdown complete")
	}

	return result
}

func (s *Server) registerGateway() error {
	_, port, err := net.SplitHostPort(s.Config.API.GRPC.ListenAddress)
	if err != nil {
		return errors.Wrap(err, "failed to determine port from configured GRPC listen address")
	}

	dialAddr := fmt.Sprintf("dns:///127.0.0.1:%s", port)

	tlsCreds, err := certs.GatewayAsClientTLSCreds(s.Config.API.GRPC.Certs)
	if err != nil {
		return errors.Wrap(err, "failed to calculate tls config for gateway service")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithTimeout(2 * time.Second), // nolint:staticcheck // using context.WithTimeout makes us unable to call defer ctx.Cancel
	}

	err = s.handlerRegistrations(s.Context, s.gtwMux, dialAddr, opts)
	if err != nil {
		return errors.Wrap(err, "failed to register handlers with the gateway")
	}

	return nil
}

func (s *Server) startHealthService(listenAddress string) error {
	healthListener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		s.logger.Error().Err(err).Str("address", listenAddress).Msg("grpc health socket failed to listen")
		return errors.Wrap(err, "grpc health socket failed to listen")
	}

	s.logger.Info().Str("address", listenAddress).Msg("GRPC Health Server starting")
	s.ErrGroup.Go(func() error {
		return s.healthServer.GRPCServer.Serve(healthListener)
	})

	return nil
}

func (s *Server) startGRPCServer(listenAddress string) error {
	s.logger.Info().Str("address", listenAddress).Msg("GRPC Server starting")
	grpcListener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return errors.Wrap(err, "grpc socket failed to listen")
	}

	s.ErrGroup.Go(func() error {
		err := s.grpcServer.Serve(grpcListener)
		if err != nil {
			s.logger.Error().Err(err).Str("address", listenAddress).Msg("GRPC Server failed to listen")
		}
		return errors.Wrap(err, "grpc server failed to listen")
	})

	return nil
}

func (s *Server) startGatewayServer(listenAddress string) error {
	s.logger.Info().Msg("Registering OpenAPI Gateway handlers")
	if err := s.registerGateway(); err != nil {
		return errors.Wrap(err, "failed to register grpc gateway handlers")
	}

	s.logger.Info().
		Str("address", "https://"+listenAddress).
		Msg("gRPC-Gateway and OpenAPI endpoint starting")
	s.ErrGroup.Go(func() error {
		return s.gtwServer.ListenAndServeTLS("", "")
	})

	return nil
}

func (s *Server) stopHTTPServer(srv *http.Server) error {
	ctx, shutdownCancel := context.WithTimeout(s.Context, 5*time.Second)
	defer shutdownCancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		if err == context.Canceled {
			s.logger.Info().Msg("server context was canceled - shutting down")
		} else {
			return err
		}
	}

	return nil
}
