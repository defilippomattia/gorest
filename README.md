# gorest

Template for a REST API in Go. Currently everything is in the main.go file, but it will be split in the future.

## Project start

```
go mod init github.com/defilippomattia/gorest
go get github.com/go-playground/validator/v10
go get github.com/rs/zerolog/log
go get github.com/go-chi/chi/v5
go get github.com/danielgtaylor/huma/v2
go get github.com/jackc/pgx/v5
go run main.go /path/to/config.json

```

API docs on http://127.0.0.1:<port>/docs

## Todo

- [] AuthN
- [] AuthZ
- [x] Config validation
- [] Encrypted config values
- [x] Structured logs
- [] REST API
- [] DB connection
- [] Dockerfile
- [] REST API docs
- [] Dynamic log level change

