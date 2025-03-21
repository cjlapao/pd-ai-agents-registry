package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv  string        `mapstructure:"app_env"`
	Server  ServerConfig  `mapstructure:"server"`
	MongoDB MongoDBConfig `mapstructure:"mongodb"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	S3      S3Config      `mapstructure:"s3"`
	Admin   AdminConfig   `mapstructure:"admin"`
}

func (c *Config) GetBaseURL() string {
	url := c.Server.Host
	if c.Server.Scheme == "https" {
		url = "https://" + url
	} else {
		url = "http://" + url
	}
	if c.Server.Port != "" {
		if c.Server.Scheme == "https" && c.Server.Port != "443" {
			url = url + ":" + c.Server.Port
		} else if c.Server.Scheme == "http" && c.Server.Port != "80" {
			url = url + ":" + c.Server.Port
		}
	}

	return url
}

type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiryHour int    `mapstructure:"expiry_hour"`
}

type S3Config struct {
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Endpoint        string `mapstructure:"endpoint"`
}

type AdminConfig struct {
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Port   string `mapstructure:"port"`
	Host   string `mapstructure:"host"`
	Scheme string `mapstructure:"scheme"`
}

func Load() (*Config, error) {
	// Reset Viper to ensure clean state
	viper.Reset()
	v := viper.New()

	// Set up Viper
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../")

	// Enable environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("")

	// First read the config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Printf("No config file found, using environment variables")
	}

	// Explicitly bind environment variables with underscores
	if err := v.BindEnv("server.port", "SERVER__PORT"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("server.host", "SERVER__HOST"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("server.scheme", "SERVER__SCHEME"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("s3.bucket", "S3__BUCKET"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("s3.access_key_id", "S3__ACCESS_KEY_ID"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("s3.secret_access_key", "S3__SECRET_ACCESS_KEY"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("s3.endpoint", "S3__ENDPOINT"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("s3.region", "S3__REGION"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("mongodb.uri", "MONGODB__URI"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("mongodb.database", "MONGODB__DATABASE"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("jwt.secret", "JWT__SECRET"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("jwt.expiry_hour", "JWT__EXPIRY_HOUR"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}
	if err := v.BindEnv("admin.password", "ADMIN__PASSWORD"); err != nil {
		return nil, fmt.Errorf("error binding environment variable: %w", err)
	}

	// Create config struct
	config := &Config{}

	// Debug print all viper settings before unmarshaling
	log.Printf("Viper settings before unmarshaling:")
	for _, key := range v.AllKeys() {
		log.Printf("%s: %v", key, v.Get(key))
	}

	// Unmarshal config
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Debug print the config struct after unmarshaling
	log.Printf("Config struct after unmarshaling:")
	log.Printf("port=%s, S3 Config: Region=%s, Bucket=%s, AccessKeyID=%s, SecretAccessKey=%s, Endpoint=%s",
		config.Server.Port,
		config.S3.Region,
		config.S3.Bucket,
		config.S3.AccessKeyID,
		config.S3.SecretAccessKey,
		config.S3.Endpoint,
	)

	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}

	// Validate required fields
	if config.S3.Region == "" {
		return nil, fmt.Errorf("S3__REGION is required")
	}
	if config.S3.AccessKeyID == "" {
		return nil, fmt.Errorf("S3__ACCESS_KEY_ID is required")
	}
	if config.S3.SecretAccessKey == "" {
		return nil, fmt.Errorf("S3__SECRET_ACCESS_KEY is required")
	}
	if config.S3.Bucket == "" {
		return nil, fmt.Errorf("S3__BUCKET is required")
	}
	if config.Admin.Password == "" {
		return nil, fmt.Errorf("ADMIN__PASSWORD is required")
	}

	// Print final config
	log.Printf("Loaded configuration: AppEnv=%s, Region=%s, Endpoint=%s, Bucket=%s",
		config.AppEnv,
		config.S3.Region,
		config.S3.Endpoint,
		config.S3.Bucket,
	)

	return config, nil
}
