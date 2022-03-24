package testharness

import (
	"path/filepath"
	"runtime"

	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
)

// AssetsDir returns the directory containing test assets
func AssetsDir() string {
	_, filename, _, _ := runtime.Caller(0) // nolint:dogsled

	return filepath.Join(filepath.Dir(filename), "testdata")
}

// AssetDefaultConfig returns the path of the default yaml config file
func AssetDefaultConfig() config.Path {
	return config.Path(filepath.Join(AssetsDir(), "config.yaml"))
}
