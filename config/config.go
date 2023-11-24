package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config - App config
type Config struct {
	Secret      Secret   `yaml:"secret"`
	Server      Server   `yaml:"server"`
	DatabaseCfg DBConfig `yaml:"db"`
}

type Secret struct {
	Key string `yaml:"key"`
}

type Server struct {
	Listen string `yaml:"listen"`
	Port   string `yaml:"port"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// LoadConfig - load config from yml file to struct
func LoadConfig(filename string) (*Config, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(f, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
