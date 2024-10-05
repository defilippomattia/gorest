package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
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

type HealthCheckOutput struct {
	Body struct {
		Message string `json:"message" example:"OK" doc:"Health status"`
	}
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

	printConfig(config)

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		//should never happen as we have already validated the log level
		log.Error().Err(err).Msg("error parsing log level, exiting application...")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(logLevel)

	dbConnURL := "postgres://" + config.Database.Username + ":" + config.Database.Password + "@" + config.Database.Host + ":" + config.Database.Port + "/" + config.Database.Name
	//"postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(context.Background(), dbConnURL)
	if err != nil {
		log.Error().Err(err).Msg("unable to connect to database, exiting application...")
		os.Exit(1)
	}
	log.Info().Msg("connected to database successfully")
	defer conn.Close(context.Background())
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("gorest API", "1.0.0"))

	huma.Get(api, "/api/healthz", func(ctx context.Context, input *struct {
	}) (*HealthCheckOutput, error) {
		resp := &HealthCheckOutput{}
		resp.Body.Message = "OK"
		return resp, nil
	})

	apiEndpoint := "127.0.0.1:" + config.APIPort

	log.Info().Msg("API server is listening on  " + apiEndpoint)

	http.ListenAndServe(apiEndpoint, router)

}
