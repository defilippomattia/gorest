package config

import (
	"encoding/json"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type Config struct {
	LogLevel string `json:"log_level" validate:"oneof=panic fatal error warn info debug trace"`
	APIPort  string `json:"api_port" validate:"required"`
	Database struct {
		Host     string `json:"host" validate:"required"`
		Port     string `json:"port" validate:"required"`
		Name     string `json:"name" validate:"required"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"database"`
}

func printConfig(config Config) {
	log.Info().
		Str("log_level", config.LogLevel).
		Str("api_port", config.APIPort).
		Str("database.host", config.Database.Host).
		Str("database.port", config.Database.Port).
		Str("database.name", config.Database.Name).
		Str("database.username", config.Database.Username).
		Str("database.password", "************").
		Msg("")
}

func validateConfig(config *Config) error {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			log.Error().
				Str("field", fieldErr.Field()).
				Str("tag", fieldErr.Tag()).
				Msg("validation error")
		}
		return err
	}
	return nil
}

func ReadConfig(configFilePath string) (*Config, error) {
	var config Config
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Error().Err(err).Msg("error reading config file")
		return nil, err
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Error().Err(err).Msg("error unmarshalling config file")
		return nil, err
	}

	printConfig(config)

	err = validateConfig(&config)
	if err != nil {
		log.Error().Err(err)
		return nil, err
	}
	log.Info().Msg("config file is valid")

	return &config, nil
}
