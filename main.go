package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/argon2"
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

type Employee struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

type EmployeesOutput struct {
	Body struct {
		Employees []Employee `json:"employees"`
	}
}

type EmployeeOutput struct {
	Body Employee `json:"body"`
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

func hashPassword(plainPassword string) (hashedPassword string, err error) {

	memory := uint32(64 * 1024)
	iterations := uint32(3)
	parallelism := uint8(2)
	saltLength := uint32(16)
	keyLength := uint32(32)
	salt := make([]byte, saltLength)
	// fmt.Printf("salt1: %v\n", salt)
	rand.Read(salt)
	// fmt.Printf("salt2: %v\n", salt)
	hashBytes := argon2.IDKey([]byte(plainPassword), salt, iterations, memory, parallelism, keyLength)
	// fmt.Printf("hashBytes: %v\n)", hashBytes)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hashBytes)
	hashedPassword = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, iterations, parallelism, b64Salt, b64Hash)
	// fmt.Printf("hashedPassword: %v\n", hashedPassword)
	return hashedPassword, nil

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

	huma.Get(api, "/api/employees", func(ctx context.Context, input *struct {
	}) (*EmployeesOutput, error) {
		log.Info().
			Str("event", "get.employees").
			Msg("getting all employees started")
		rows, err := conn.Query(context.Background(), "SELECT id, first_name, last_name, email, age, created_at FROM employees")
		if err != nil {
			log.Error().
				Str("event", "get.employees").
				Err(err).Msg("error fetching employees")
			return nil, err
		}
		defer rows.Close()
		var employees []Employee
		for rows.Next() {
			var employee Employee
			err = rows.Scan(&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email, &employee.Age, &employee.CreatedAt)
			if err != nil {
				log.Error().Err(err).Msg("error scanning employee")
				return nil, err
			}
			employees = append(employees, employee)
		}
		resp := &EmployeesOutput{}
		resp.Body.Employees = employees
		return resp, nil
	})

	huma.Get(api, "/api/employees/{id}", func(ctx context.Context, input *struct {
		ID int `path:"id"`
	}) (*EmployeeOutput, error) {
		row := conn.QueryRow(context.Background(), "SELECT id, first_name, last_name, email, age, created_at FROM employees WHERE id = $1", input.ID)

		var employee Employee
		err := row.Scan(&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email, &employee.Age, &employee.CreatedAt)
		if err != nil {
			if err == pgx.ErrNoRows {
				log.Error().Err(err).Msg("employee not found")
				return nil, err
			}
			log.Error().Err(err).Msg("error fetching employee")
			return nil, err
		}

		resp := &EmployeeOutput{
			Body: employee,
		}
		return resp, nil
	})

	huma.Post(api, "/api/employees", func(ctx context.Context, input *struct {
		Body struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Age       int    `json:"age"`
		} `json:"body"`
	}) (*EmployeeOutput, error) {
		var employeeID int
		err := conn.QueryRow(context.Background(),
			"INSERT INTO employees (first_name, last_name, email, age) VALUES ($1, $2, $3, $4) RETURNING id",
			input.Body.FirstName, input.Body.LastName, input.Body.Email, input.Body.Age).Scan(&employeeID)

		if err != nil {
			log.Error().Err(err).Msg("error inserting new employee")
			return nil, err
		}
		row := conn.QueryRow(context.Background(),
			"SELECT id, first_name, last_name, email, age, created_at FROM employees WHERE id = $1", employeeID)
		var employee Employee
		err = row.Scan(&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email, &employee.Age, &employee.CreatedAt)
		if err != nil {
			log.Error().Err(err).Msg("error fetching newly created employee")
			return nil, err
		}
		resp := &EmployeeOutput{
			Body: employee,
		}
		return resp, nil
	})

	huma.Delete(api, "/api/employees/{id}", func(ctx context.Context, input *struct {
		ID int `path:"id"`
	}) (*struct{}, error) {
		_, err := conn.Exec(context.Background(), "DELETE FROM employees WHERE id = $1", input.ID)
		if err != nil {
			log.Error().Err(err).Msg("error deleting employee")
			return nil, err
		}
		return nil, nil
	})

	huma.Post(api, "/api/users", func(ctx context.Context, input *struct {
		Body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"body"`
	}) (*UserOutput, error) {
		// var username string
		hashedPassword, err := hashPassword(input.Body.Password)
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

	apiEndpoint := "127.0.0.1:" + config.APIPort

	log.Info().Msg("API server is listening on  " + apiEndpoint)

	http.ListenAndServe(apiEndpoint, router)

}
