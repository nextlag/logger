package configuration

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

var (
	cfg        Config
	once       sync.Once
	configPath string
)

func init() {
	configPath = os.Getenv("CONFIG_PATH")
}

type Config struct {
	Logging *Logging `yaml:"logging"`
}

type Logging struct {
	Level       slog.Level `yaml:"level" env:"LOG_LEVEL" envDefault:"debug"`
	ProjectPath string     `yaml:"project_path" env:"PROJECT_PATH"`
	LogToFile   bool       `yaml:"log_to_file" env:"LOG_TO_FILE" envDefault:"false"`
	LogPath     string     `yaml:"log_path" env:"LOG_PATH" envDefault:"/Users/nextbug/Documents/GoProjects/concurrency/data/logs/out.log"`
}

func Load() (*Config, error) {
	once.Do(func() {
		if configPath != "" {
			err := loadConfigFromYAML(&cfg)
			if err != nil {
				return
			}
		}

		flag.BoolVar(&cfg.Logging.LogToFile, "out_log", cfg.Logging.LogToFile, "log output to file")
		flag.StringVar(&cfg.Logging.LogPath, "dir_log", cfg.Logging.LogPath, "path to the log file")
		flag.Var(&LogLevelValue{&cfg.Logging.Level}, "level", "Log level (debug, info, warn, error)")
		flag.StringVar(&cfg.Logging.ProjectPath, "project", cfg.Logging.ProjectPath, "Path to the current project")
		flag.Parse()

		if err := env.Parse(&cfg); err != nil {
			return
		}
	})

	return &cfg, nil
}

func loadConfigFromYAML(cfg *Config) error {
	if configPath == "" {
		log.Println("the path to the configuration file is empty")
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	return nil
}
