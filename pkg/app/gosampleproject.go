package app

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/aserto-dev/go-sample-project/pkg/app/server"
	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
)

type GoSampleProject struct {
	Context       context.Context
	Logger        *zerolog.Logger
	Configuration *config.Config
	Server        *server.Server
}
