package main

import (
	"encoding/json"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	LogLevel string `json:"log_level" validate:"oneof=panic fatal error warn info debug trace"`
	Database struct {
		Host     string `json:"host" validate:"required"`
		Port     string `json:"port" validate:"required"`
		Name     string `json:"name" validate:"required"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"database"`
}

func main() {

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	//until the log level is set from config file, set it to trace
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	if len(os.Args) != 2 {
		log.Error().Msg("provide config file path, example: go run main.go /path/to/config.prod.json")
		os.Exit(1)
	}
	configFilePath := os.Args[1]
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Error().Err(err).Msg("error reading config file, exiting application...")
		os.Exit(1)
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Error().Err(err).Msg("error unmarshalling config file, exiting application...")
		os.Exit(1)
	}
	log.Info().Msg("config file " + configFilePath + " read successfully")
	validate := validator.New()
	err = validate.Struct(config)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			log.Error().
				Str("field", fieldErr.Field()).
				Str("tag", fieldErr.Tag()).
				Msg("validation error")
		}
		log.Error().Msg("exiting application...")
		os.Exit(1)
	}

	log.Info().Msg("config file is valid")

	//todo: print config values

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		//should never happen as we have already validated the log level
		log.Error().Err(err).Msg("error parsing log level, exiting application...")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(logLevel)
}
