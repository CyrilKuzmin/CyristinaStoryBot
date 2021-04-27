package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type (
	Config struct {
		TG    TgConfig
		Mongo MongoConfig
	}

	TgConfig struct {
		Token          string
		NextMsgTimeout time.Duration
		ImagesDir      string
	}

	MongoConfig struct {
		URI        string
		Database   string
		Collection string
	}
)

const (
	defaultNextMsgTimeout  = time.Minute * 15
	defaultMongoURI        = "mongodb://localhost:27017"
	defaultMongoDatabase   = "stories"
	defaultMongoCollection = "stroies"
)

func Init(path string) (*Config, error) {
	setDefaults()

	if err := parseEnv(); err != nil {
		return nil, err
	}

	if err := parseConfigFile(path); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("Mongo.URI", defaultMongoURI)
	viper.SetDefault("Mongo.Database", defaultMongoDatabase)
	viper.SetDefault("Mongo.Collection", defaultMongoCollection)
	viper.SetDefault("TG.NextMsgTimeout", defaultNextMsgTimeout)
}

func parseConfigFile(filepath string) error {
	path := strings.Split(filepath, "/")

	viper.AddConfigPath(path[0]) // folder
	viper.SetConfigName(path[1]) // config file name

	return viper.ReadInConfig()
}

func parseEnv() error {
	if err := parseMongoEnvVariables(); err != nil {
		return err
	}

	if err := parseTGFromEnv(); err != nil {
		return err
	}

	return nil
}

func parseMongoEnvVariables() error {
	viper.SetEnvPrefix("mongo")
	return viper.BindEnv("URI")
}

func parseTGFromEnv() error {
	viper.SetEnvPrefix("tg")
	return viper.BindEnv("token")
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("TG", &cfg.TG); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("Mongo", &cfg.Mongo); err != nil {
		return err
	}

	return nil
}

func setFromEnv(cfg *Config) {
	cfg.Mongo.URI = viper.GetString("URI")
	cfg.TG.Token = viper.GetString("Token")
}
