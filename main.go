package main

import (
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/defilippomattia/gorest/apis/companies"
	"github.com/defilippomattia/gorest/apis/users"
	"github.com/defilippomattia/gorest/config"
	"github.com/defilippomattia/gorest/database"
	"github.com/defilippomattia/gorest/employees"
	"github.com/defilippomattia/gorest/healthz"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	//until the log level is set from config file, set it to trace
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	if len(os.Args) != 2 {
		log.Error().Msg("provide config file path, example: go run main.go /path/to/config.prod.json")
		os.Exit(1)
	}
	configFilePath := os.Args[1]
	cfg, err := config.ReadConfig(configFilePath)
	if err != nil {
		log.Info().Msg("exiting application...")
		os.Exit(1)
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		//should never happen as we have already validated the log level
		log.Error().Err(err).Msg("error parsing log level, exiting application...")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(logLevel)

	dbConnURL := "postgres://" + cfg.Database.Username + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + cfg.Database.Port + "/" + cfg.Database.Name
	conn, err := database.ConnectToDatabase(dbConnURL)
	if err != nil {
		log.Error().Msg("exiting application...")
		os.Exit(1)
	}
	defer conn.Close()

	log.Info().Msg("connected to database successfully")
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("gorest API", "1.0.0"))

	huma.Get(api, "/api/healthz", healthz.GetHealth)

	huma.Get(api, "/api/employees", employees.GetEmployees(conn))
	huma.Get(api, "/api/employees/{id}", employees.GetEmployeeById(conn))
	huma.Post(api, "/api/employees", employees.CreateEmployee(conn))

	companyRepo := companies.NewPgCompanyRepository(conn)
	companyHandler := companies.NewCompanyHandler(companyRepo)

	router.Post("/api/companies", companyHandler.CreateCompany)
	router.Get("/api/companies", companyHandler.GetCompanies)
	router.Get("/api/companies/{id}", companyHandler.GetCompanyByID)

	userRepo := users.NewPgUserRepository(conn)
	userHandler := users.NewUserHandler(userRepo)

	router.Post("/api/register", userHandler.RegisterUser)

	apiEndpoint := "127.0.0.1:" + cfg.APIPort

	log.Info().Msg("API server is listening on  " + apiEndpoint)

	http.ListenAndServe(apiEndpoint, router)

}
