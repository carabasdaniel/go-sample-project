package main

import (
	"os"

	"github.com/alecthomas/kong"

	"github.com/aserto-dev/go-sample-project/pkg/app"
	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
	"github.com/aserto-dev/go-sample-project/pkg/version"
)

type RunCmd struct {
}

func (r *RunCmd) Run(globals *Globals) error {
	configFile := globals.Config

	appInstance, cleanup, err := app.BuildGoSampleProject(
		os.Stdout,
		os.Stderr,
		config.Path(configFile),
		func(*config.Config) {})

	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()
	if err != nil {
		return err
	}

	err = appInstance.Server.Start()
	if err != nil {
		return err
	}

	<-appInstance.Context.Done()

	return nil
}

type VersionCmd struct {
}

func (cmd *VersionCmd) Run(globals *Globals) error {
	configFile := globals.Config

	appInstance, cleanup, err := app.BuildGoSampleProject(
		os.Stdout,
		os.Stderr,
		config.Path(configFile),
		func(*config.Config){})

	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()
	if err != nil {
		return err
	}

	appInstance.Logger.Info().
		Str("version",  version.GetInfo().Version).
		Str("date",  version.GetInfo().Date).
		Str("commit",  version.GetInfo().Commit).
		Msg("go Sample Project")

	return nil
}

type Globals struct {
	Config string `short:"c" help:"path of the configuration file" default:"config.yaml"`
}

type CLI struct {
	Globals
	Run RunCmd `cmd:"" help:"Run go Sample Project service"`
	Version VersionCmd `cmd:"" help:"Print version and exit"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli, kong.UsageOnError())
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
