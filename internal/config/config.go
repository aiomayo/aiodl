package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/aiomayo/aiodl/internal/paths"
)

const (
	FileName = "config.yaml"
)

type Config struct {
	DownloadDir  string `mapstructure:"download_dir"`
	OutputFormat string `mapstructure:"output_format"`
	Quality      string `mapstructure:"quality"`
	Verbose      bool   `mapstructure:"verbose"`
	Parallel     int    `mapstructure:"parallel"`
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("download_dir", paths.DownloadDir())
	v.SetDefault("output_format", "mp4")
	v.SetDefault("quality", "best")
	v.SetDefault("verbose", false)
	v.SetDefault("parallel", 3)
}

func Load(v *viper.Viper) (*Config, error) {
	SetDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(paths.ConfigDir())
	v.AddConfigPath(".")
	v.SetEnvPrefix("AIODL")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	return &cfg, nil
}

func FilePath() string {
	path, _ := paths.ConfigFile(FileName)
	return path
}
