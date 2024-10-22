package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/defilippomattia/gorest/auth"
	"github.com/defilippomattia/gorest/employees"
	"github.com/defilippomattia/gorest/healthz"
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

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

type UserOutput struct {
	Body UserResponse `json:"body"`
}
type HealthCheckOutput struct {
	Body struct {
		Message string `json:"message" example:"OK" doc:"Health status"`
	}
}
type LoginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Token    string `json:"token"`
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
	} `json:"body"`
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
	conn, err := pgx.Connect(context.Background(), dbConnURL)
	if err != nil {
		log.Error().Err(err).Msg("unable to connect to database, exiting application...")
		os.Exit(1)
	}
	log.Info().Msg("connected to database successfully")
	defer conn.Close(context.Background())
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("gorest API", "1.0.0"))

	huma.Get(api, "/api/healthz", healthz.GetHealth)

	huma.Get(api, "/api/employees", employees.GetEmployees(conn))
	huma.Get(api, "/api/employees/{id}", employees.GetEmployeeById(conn))
	huma.Post(api, "/api/employees", employees.CreateEmployee(conn))

	huma.Post(api, "/api/users", func(ctx context.Context, input *struct {
		Body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"body"`
	}) (*UserOutput, error) {
		// var username string
		hashedPassword, err := auth.HashPassword(input.Body.Password)
		if err != nil {
			log.Error().Err(err).Msg("error hashing password")
			return nil, err
		}
		var userId int
		err = conn.QueryRow(context.Background(),
			"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
			input.Body.Username, hashedPassword).Scan(&userId)
		if err != nil {
			log.Error().Err(err).Msg("error inserting new employee")
			return nil, err
		}
		row := conn.QueryRow(context.Background(),
			"SELECT id,username,password FROM users WHERE id = $1", userId)
		var user User
		err = row.Scan(&user.Id, &user.Username, &user.Password)
		if err != nil {
			log.Error().Err(err).Msg("error fetching newly created user")
			return nil, err
		}
		resp := &UserOutput{
			Body: UserResponse{
				Id:       userId,
				Username: input.Body.Username,
			},
		}
		return resp, nil
	})

	huma.Post(api, "/api/auth/login", func(ctx context.Context, input *struct {
		Body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"body"`
	}) (*LoginOutput, error) {
		hashedPassword, err := auth.HashPassword(input.Body.Password)
		if err != nil {
			log.Error().Err(err).Msg("error hashing password")
			return nil, err
		}
		fmt.Printf("hashedPassword: %v\n", hashedPassword)

		var user User
		err = conn.QueryRow(context.Background(),
			"SELECT id, username, password FROM users WHERE username = $1", input.Body.Username).Scan(&user.Id, &user.Username, &user.Password)
		if err != nil {
			log.Error().Err(err).Msg("error fetching user")
			return nil, err
		}
		//todo: compare hashedPassword with user.Password
		token := auth.GenerateSessionToken()
		currentTimestamp := time.Now()
		_, err = conn.Exec(context.Background(),
			"INSERT INTO sessions (token, user_id, created_at, last_used) VALUES ($1, $2, $3, $4)", token, user.Id, currentTimestamp, currentTimestamp)
		if err != nil {
			log.Error().Err(err).Msg("error inserting new session")
			return nil, err
		}

		expiresAt := currentTimestamp.Add(1 * time.Minute)
		resp := &LoginOutput{
			SetCookie: http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  expiresAt,
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
			},
			Body: struct {
				Token    string `json:"token"`
				UserID   int    `json:"user_id"`
				Username string `json:"username"`
			}{
				Token:    token,
				UserID:   user.Id,
				Username: user.Username,
			},
		}

		return resp, nil
	})

	apiEndpoint := "127.0.0.1:" + config.APIPort

	log.Info().Msg("API server is listening on  " + apiEndpoint)

	http.ListenAndServe(apiEndpoint, router)

}
