package peer

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"

	"github.com/daeMOn63/gorrent/fs"
)

// Config list the options for the Peer configuration
type Config struct {
	ID         string `json:"id"`
	PublicIP   net.IP `json:"publicIP"`
	PublicPort uint16 `json:"publicPort"`
	SockPath   string `json:"socketPath"`
	DbPath     string `json:"dbPath"`
}

// Configurator allow to load a configuration
type Configurator interface {
	Load(path string) (*Config, error)
}

// ConfigValidator validates given configuration
type ConfigValidator interface {
	Validate(cfg *Config) error
}

type configValidator struct {
}

var _ Configurator = &jsonConfigurator{}
var _ ConfigValidator = &configValidator{}

type jsonConfigurator struct {
	fs        fs.FileSystem
	validator ConfigValidator
}

// NewJSONConfigurator creates a new Configurator
func NewJSONConfigurator(fs fs.FileSystem, validator ConfigValidator) Configurator {
	return &jsonConfigurator{
		fs:        fs,
		validator: validator,
	}
}

// Load read a json configuration from path, unmarshall and returns it
func (c *jsonConfigurator) Load(path string) (*Config, error) {
	file, err := c.fs.Open(path)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(buf.Bytes(), config)
	if err != nil {
		return nil, err
	}

	if err := c.validator.Validate(config); err != nil {
		return nil, err
	}

	return config, nil
}

// NewConfigValidator returns a new configuration validator
func NewConfigValidator() ConfigValidator {
	return &configValidator{}
}

// Validation errors
var (
	ErrConfigIDRequired       = errors.New("config: id is required")
	ErrConfigSockPathRequired = errors.New("config: socketPath is required")
	ErrConfigDbPathRequired   = errors.New("config: dbPath is required")
)

// Validate check given configuration and returns errors when any fields has invalid value
func (c *configValidator) Validate(cfg *Config) error {

	if len(cfg.ID) == 0 {
		return ErrConfigIDRequired
	}

	if len(cfg.SockPath) == 0 {
		return ErrConfigSockPathRequired
	}

	if len(cfg.DbPath) == 0 {
		return ErrConfigDbPathRequired
	}

	return nil
}
