//+build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/aserto-dev/go-utils/logger"
	"github.com/aserto-dev/go-sample-project/pkg/app/impl"
	"github.com/aserto-dev/go-sample-project/pkg/app/server"
	"github.com/aserto-dev/go-sample-project/pkg/cc"
	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
)

var (
	gosampleprojectSet = wire.NewSet(
		cc.NewCC,

		GRPCServerRegistrations,
		GatewayServerRegistrations,
		server.NewServer,

		impl.NewInfo,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)

	gosampleprojectTestSet = wire.NewSet(
		// Test
		cc.NewTestCC,

		// Normal
		GRPCServerRegistrations,
		GatewayServerRegistrations,
		server.NewServer,

		impl.NewInfo,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)
)

func BuildGoSampleProject(
	logWriter logger.Writer,
	errWriter logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (*GoSampleProject, func(), error) {
	wire.Build(
		wire.Struct(new(GoSampleProject), "*"),
		gosampleprojectSet,
	)
	return &GoSampleProject{}, func() {}, nil
}

func BuildTestGoSampleProject(
	logWriter logger.Writer,
	errWriter logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (*GoSampleProject, func(), error) {
	wire.Build(
		wire.Struct(new(GoSampleProject), "*"),
		gosampleprojectTestSet,
	)
	return &GoSampleProject{}, func() {}, nil
}
