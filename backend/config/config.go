package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 是整个应用的配置根结构。
type Config struct {
	Server ServerConfig `yaml:"server"`
	App    AppConfig    `yaml:"app"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// Global 是加载后的全局配置，便于其他包直接引用。
var Global *Config

// Load 从指定路径读取并解析 yaml 配置。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	Global = cfg
	return cfg, nil
}
