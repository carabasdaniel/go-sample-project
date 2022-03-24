package impl

import (
	"context"
	"runtime"

	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
	"github.com/aserto-dev/go-sample-project/pkg/version"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
	"github.com/rs/zerolog"
)

// Info is an implementation of the info API
type Info struct {
	logger  *zerolog.Logger
	cfg     *config.Config
}

// NewInfo creates a new Info
func NewInfo(logger *zerolog.Logger, cfg *config.Config) *Info {
	serviceLogger := logger.With().Str("component", "impl.go Sample Project").Logger()

	return &Info{
		logger:  &serviceLogger,
		cfg:     cfg,
	}
}

func (i *Info) Info(context.Context, *info.InfoRequest) (*info.InfoResponse, error) {
	buildVersion := version.GetInfo()

	return &info.InfoResponse{
		Build: &info.BuildInfo{
			Version: buildVersion.Version,
			Commit:  buildVersion.Commit,
			Date:    buildVersion.Date,
			Os:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
	}, nil
}
