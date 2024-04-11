package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		HTTP   `yaml:"http"`
		Logger `yaml:"logger"`
		Auth   `yaml:"auth"`
		DB     `yaml:"database"`
	}

	HTTP struct {
		Port int `yaml:"port"`
	}

	Logger struct {
		Level string `yaml:"level"`
	}

	Auth struct {
		TokenTTL   time.Duration `yaml:"token_ttl"`
		SignSecret string        `yaml:"sign_secret"`
	}

	DB struct {
		Url string `yaml:"url"`
	}
)

func NewConfig() (*Config, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("config geting executable filepath error: %w", err)
	}
	yamlFile, err := os.ReadFile(filepath.Join(executablePath, "../../config.yml"))
	if err != nil {
		return nil, fmt.Errorf("config reading yaml file error: %w", err)
	}

	// Default values
	config := Config{
		HTTP: HTTP{
			Port: 3000,
		},
		Logger: Logger{
			Level: "debug",
		},
		Auth: Auth{
			TokenTTL: 30 * time.Minute,
		},
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, fmt.Errorf("config parsing yaml file error: %w", err)
	}

	// Parse environment variables
	httpPort, ok := os.LookupEnv("HTTP_PORT")
	if ok {
		httpPortInt, err := strconv.Atoi(httpPort)
		if err != nil {
			return nil, fmt.Errorf("environment variable HTTP_PORT converting error: %w", err)
		}
		config.HTTP.Port = httpPortInt
	}

	loggerLevel, ok := os.LookupEnv("LOGGER_LEVEL")
	if ok {
		config.Logger.Level = loggerLevel
	}

	authTokenTTL, ok := os.LookupEnv("AUTH_TOKEN_TTL")
	if ok {
		tokenTTLParsed, err := time.ParseDuration(authTokenTTL)
		if err != nil {
			return nil, fmt.Errorf("environment variable AUTH_TOKEN_TTL parsing error: %w", err)
		}
		config.Auth.TokenTTL = tokenTTLParsed
	}

	authSignSecret, ok := os.LookupEnv("AUTH_SIGN_SECRET")
	if ok {
		config.Auth.SignSecret = authSignSecret
	}

	dbUrl, ok := os.LookupEnv("DB_URL")
	if ok {
		config.DB.Url = dbUrl
	}

	return &config, nil
}
