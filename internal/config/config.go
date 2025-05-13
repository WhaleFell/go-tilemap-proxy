package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ref: https://github.com/spf13/viper?tab=readme-ov-file#unmarshaling

type ServerConfig struct {
	Host string `json:"host" yaml:"host" mapstructure:"host"`
	Port int    `json:"port" yaml:"port" mapstructure:"port" `
}

type CacheConfig struct {
	Enable bool   `json:"enable" yaml:"enable" mapstructure:"enable"`
	Path   string `json:"path" yaml:"path" mapstructure:"path"`
	MaxAge int    `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
}

type LogConfig struct {
	Level      string `json:"level" yaml:"level" mapstructure:"level"`
	EnableFile bool   `json:"enable_file" yaml:"enable_file" mapstructure:"enable_file"`
	FilePath   string `json:"file_path" yaml:"file_path" mapstructure:"file_path"`
}

type Config struct {
	Server ServerConfig `json:"server" yaml:"server" mapstructure:"server"`
	Cache  CacheConfig  `json:"cache" yaml:"cache" mapstructure:"cache"`
	Log    LogConfig    `json:"log" yaml:"log" mapstructure:"log"`
	Proxy  string       `json:"proxy" yaml:"proxy" mapstructure:"proxy"`
}

var Cfg *Config

func InitConfig(configPath string) error {
	// viper can recognize config file type automatically.

	// set config type
	// viper.SetConfigType("yaml")
	// viper.SetConfigType("json")
	// viper.SetConfigType("toml")
	// viper.SetConfigFile("env")

	// set default server config
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8076)

	// set default cache config
	viper.SetDefault("cache.path", "./cache")
	viper.SetDefault("cache.enable", true)
	viper.SetDefault("cache.max_age", 3600)

	// set default log config
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.enable_file", true)
	viper.SetDefault("log.file_path", "")

	// set default proxy config
	viper.SetDefault("proxy", "")

	// load config in environment variable
	viper.AutomaticEnv()

	// read config file
	file, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		// create default config file
		if err := viper.SafeWriteConfigAs(configPath); err != nil {
			fmt.Printf("create default config file %s failed: %v\n", configPath, err)
		}
		return fmt.Errorf("open config file %s failed: %w, create default config now", configPath, err)
	}

	// get file extension name and set config type
	ext := filepath.Ext(configPath)
	viper.SetConfigType(ext[1:])

	// load config file
	if err := viper.ReadConfig(file); err != nil {
		return fmt.Errorf("viper load config file %s failed: %w", configPath, err)
	}

	var conf Config

	if err := viper.Unmarshal(&conf); err != nil {
		return fmt.Errorf("viper unmarshal config file %s failed: %w", configPath, err)
	}

	Cfg = &conf
	return nil

}
