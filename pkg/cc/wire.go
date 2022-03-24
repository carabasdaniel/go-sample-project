//+build wireinject

package cc

import (
	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/go-utils/logger"
	"github.com/google/wire"

	cc_context "github.com/aserto-dev/go-sample-project/pkg/cc/context"
	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
	"github.com/aserto-dev/go-sample-project/pkg/cc/metrics"
)

var (
	ccSet = wire.NewSet(
		cc_context.NewContext,
		config.NewConfig,
		config.NewLoggerConfig,
		logger.NewLogger,
		metrics.NewPrometheusRecorder,
		certs.NewGenerator,
		wire.FieldsOf(new(config.Config), "Logging"),
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup"),

		wire.Struct(new(CC), "*"),
	)

	ccTestSet = wire.NewSet(
		// Test
		cc_context.NewTestContext,

		// Normal
		config.NewConfig,
		config.NewLoggerConfig,
		logger.NewLogger,
		metrics.NewPrometheusRecorder,
		certs.NewGenerator,
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup"),

		wire.Struct(new(CC), "*"),
	)
)

// buildCC sets up the CC struct that contains all dependencies that
// are cross cutting
func buildCC(
	logOutput logger.Writer,
	errOutput logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (*CC, func(), error) {
	wire.Build(ccSet)
	return &CC{}, func() {}, nil
}

func buildTestCC(
	logOutput logger.Writer,
	errOutput logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (*CC, func(), error) {
	wire.Build(ccTestSet)
	return &CC{}, func() {}, nil
}
