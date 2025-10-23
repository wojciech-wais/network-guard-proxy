package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy   ProxyConfig   `yaml:"proxy"`
	Admin   AdminConfig   `yaml:"admin"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
}

type ProxyConfig struct {
	ListenAddress string `yaml:"listen_address"`
	UpstreamURL   string `yaml:"upstream_url"`
}

type AdminConfig struct {
	ListenAddress string `yaml:"listen_address"`
}

type StorageConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	MaxLogs int    `yaml:"max_logs"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Proxy: ProxyConfig{
			ListenAddress: ":8080",
			UpstreamURL:   "http://localhost:9000",
		},
		Admin: AdminConfig{
			ListenAddress: ":9090",
		},
		Storage: StorageConfig{
			Driver: "sqlite",
			DSN:    "./network_guard.db",
		},
		Logging: LoggingConfig{
			Level:   "info",
			MaxLogs: 10000,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
