package server

import (
	"net/http"

	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/go-utils/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/slok/go-http-metrics/metrics"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
)

var (
	allowedOrigins = []string{
		"http://localhost",
		"http://localhost:*",
		"https://localhost",
		"https://localhost:*",
		"http://127.0.0.1",
		"http://127.0.0.1:*",
		"https://127.0.0.1",
		"https://127.0.0.1:*",
	}
)

// newGatewayServer creates a new gateway server.
func newGatewayServer(
	log *zerolog.Logger,
	cfg *config.Config,
	gtwMux *runtime.ServeMux,
	metricsRecorder metrics.Recorder,
) (*http.Server, error) {
	corsLogger := log.With().Str("source", "cors").Logger()
	gatewayLogger := log.With().Str("source", "http-gateway").Logger()

	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		Debug:          cfg.Logging.LogLevelParsed <= zerolog.DebugLevel,
	})
	c.Log = &corsLogger

	middleware := addConfiurableHandler(cfg)

	mux := http.NewServeMux()
	mux.Handle("/api/", middleware(fieldsMaskHandler(gtwMux)))

	gtwServer := &http.Server{
		ErrorLog: logger.NewSTDLogger(&gatewayLogger),
		Addr:     cfg.API.Gateway.ListenAddress,
		Handler:  c.Handler(mux),
	}

	tlsServerConfig, err := certs.GatewayServerTLSConfig(cfg.API.Gateway.Certs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate gateway server tls creds")
	}

	gtwServer.TLSConfig = tlsServerConfig

	return gtwServer, nil
}

// customHeaderMatcher is a matcher that makes it so that HTTP clients do not have to prefix
// the header key with Grpc-Metadata-.
// see https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/#mapping-from-http-request-headers-to-grpc-client-metadata
func customHeaderMatcher(key string) (string, bool) {
	switch key { // nolint:gocritic // this is a stub in the cookiecutter, remove this hint after implementing
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// fieldsMaskHandler will set the Content-Type to "application/json+masked", which
// will signal the marshaler to not emit unpopulated types, which is needed to
// serialize the masked result set.
// This happens if a fields.mask query parameter is present and set
func fieldsMaskHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := r.URL.Query()["fields.mask"]; ok && len(p) > 0 && len(p[0]) > 0 {
			r.Header.Set("Content-Type", "application/json+masked")
		}
		h.ServeHTTP(w, r)
	})
}

func addConfiurableHandler(cfg *config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler { return h }
}

// gatewayMux creates a gateway multiplexer for serving the API as an OpenAPI endpoint.
func gatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Multiline:       false,
					Indent:          "  ",
					AllowPartial:    true,
					UseProtoNames:   true,
					UseEnumNumbers:  false,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: false,
				},
			},
		),
		runtime.WithMarshalerOption(
			"application/json+masked",
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Multiline:       false,
					Indent:          "  ",
					AllowPartial:    true,
					UseProtoNames:   true,
					UseEnumNumbers:  false,
					EmitUnpopulated: false,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: false,
				},
			},
		),
	)
}
