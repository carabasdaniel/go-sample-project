package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/go-utils/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	DefaultTLSGenDir = os.ExpandEnv("$HOME/.config/aserto/go-sample-project/certs")
)

// Overrider is a func that mutates configuration
type Overrider func(*Config)

// Config holds the configuration for the app.
type Config struct {
	Logging logger.Config `json:"logging"`
	API     struct {
		GRPC struct {
			ListenAddress string `json:"listen_address"`
			// Default connection timeout is 120 seconds
			// https://godoc.org/google.golang.org/grpc#ConnectionTimeout
			ConnectionTimeoutSeconds uint32               `json:"connection_timeout_seconds"`
			Certs                    certs.TLSCredsConfig `json:"certs"`
		} `json:"grpc"`
		Gateway struct {
			ListenAddress string               `json:"listen_address"`
			Certs         certs.TLSCredsConfig `json:"certs"`
		} `json:"gateway"`
		Health struct {
			ListenAddress string `json:"listen_address"`
		} `json:"health"`
	} `json:"api"`
}

// Path is a string that points to a config file
type Path string

// NewConfig creates the configuration by reading env & files
func NewConfig(configPath Path, log *zerolog.Logger, overrides Overrider, certsGenerator *certs.Generator) (*Config, error) {
	configLogger := log.With().Str("component", "config").Logger()
	log = &configLogger

	v := viper.New()

	file := "config.yaml"
	if configPath != "" {
		exists, err := fileExists(string(configPath))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to determine if config file '%s' exists", configPath)
		}

		if !exists {
			return nil, errors.Errorf("config file '%s' doesn't exist", configPath)
		}

		file = string(configPath)
	}

	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigFile(file)
	v.SetEnvPrefix("GO_SAMPLE_PROJECT")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set defaults, e.g.
	v.SetDefault("api.grpc.certs.tls_key_path", filepath.Join(DefaultTLSGenDir, "grpc.key"))
	v.SetDefault("api.grpc.certs.tls_cert_path", filepath.Join(DefaultTLSGenDir, "grpc.crt"))
	v.SetDefault("api.grpc.certs.tls_ca_cert_path", filepath.Join(DefaultTLSGenDir, "grpc-ca.crt"))
	v.SetDefault("api.grpc.connection_timeout_seconds", 120)
	v.SetDefault("api.gateway.certs.tls_key_path", filepath.Join(DefaultTLSGenDir, "gateway.key"))
	v.SetDefault("api.gateway.certs.tls_cert_path", filepath.Join(DefaultTLSGenDir, "gateway.crt"))
	v.SetDefault("api.gateway.certs.tls_ca_cert_path", filepath.Join(DefaultTLSGenDir, "gateway-ca.crt"))
	v.SetDefault("api.grpc.listen_address", "0.0.0.0:8282")
	v.SetDefault("api.gateway.listen_address", "0.0.0.0:8383")
	v.SetDefault("api.health.listen_address", "0.0.0.0:8484")

	configExists, err := fileExists(file)
	if err != nil {
		return nil, errors.Wrapf(err, "filesystem error")
	}

	if configExists {
		if err = v.ReadInConfig(); err != nil {
			return nil, errors.Wrapf(err, "failed to read config file '%s'", file)
		}
	}
	v.AutomaticEnv()

	cfg := new(Config)

	err = v.UnmarshalExact(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config file")
	}

	if overrides != nil {
		overrides(cfg)
	}

	// This is where validation of config happens
	err = func() error {
		return nil
	}()

	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config file")
	}

	if certsGenerator != nil {
		err = cfg.setupCerts(log, certsGenerator)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup certs")
		}
	}

	return cfg, nil
}

// NewLoggerConfig creates a new LoggerConfig
func NewLoggerConfig(configPath Path, overrides Overrider) (*logger.Config, error) {
	discardLogger := zerolog.New(io.Discard)
	cfg, err := NewConfig(configPath, &discardLogger, overrides, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new config")
	}

	return &cfg.Logging, nil
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

func (c *Config) setupCerts(log *zerolog.Logger, certsGenerator *certs.Generator) error {
	existingFiles := []string{}
	for _, file := range []string{
		c.API.GRPC.Certs.TLSCACertPath,
		c.API.GRPC.Certs.TLSCertPath,
		c.API.GRPC.Certs.TLSKeyPath,
		c.API.Gateway.Certs.TLSCACertPath,
		c.API.Gateway.Certs.TLSCertPath,
		c.API.Gateway.Certs.TLSKeyPath,
	} {
		exists, err := fileExists(file)
		if err != nil {
			return errors.Wrapf(err, "failed to determine if file '%s' exists", file)
		}

		if !exists {
			continue
		}

		existingFiles = append(existingFiles, file)
	}

	if len(existingFiles) == 0 {
		err := certsGenerator.MakeDevCert(&certs.CertGenConfig{
			CommonName:       "go-sample-project-grpc",
			CertKeyPath:      c.API.GRPC.Certs.TLSKeyPath,
			CertPath:         c.API.GRPC.Certs.TLSCertPath,
			CACertPath:       c.API.GRPC.Certs.TLSCACertPath,
			DefaultTLSGenDir: DefaultTLSGenDir,
		})
		if err != nil {
			return errors.Wrap(err, "failed to generate gateway certs")
		}

		err = certsGenerator.MakeDevCert(&certs.CertGenConfig{
			CommonName:       "go-sample-project-gateway",
			CertKeyPath:      c.API.Gateway.Certs.TLSKeyPath,
			CertPath:         c.API.Gateway.Certs.TLSCertPath,
			CACertPath:       c.API.Gateway.Certs.TLSCACertPath,
			DefaultTLSGenDir: DefaultTLSGenDir,
		})
		if err != nil {
			return errors.Wrap(err, "failed to generate grpc certs")
		}
	} else {
		msg := zerolog.Arr()
		for _, f := range existingFiles {
			msg.Str(f)
		}
		log.Info().Array("existing-files", msg).Msg("some cert files already exist, skipping generation")
	}

	return nil
}
