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
go get golang.org/x/crypto/argon2

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
- [x] DB connection
- [] Dockerfile
- [] REST API docs
- [] Dynamic log level change

# DB

A demo table is created to show the API functionality. 
```sql

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,                
    first_name VARCHAR(100) NOT NULL,     
    last_name VARCHAR(100) NOT NULL,      
    email VARCHAR(255) NOT NULL UNIQUE,   
    age INT,             
    created_at TIMESTAMP DEFAULT NOW()    
);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('John', 'Doe', 'john.doe@example.com', 30);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Jane', 'Smith', 'jane.smith@example.com', 25);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Alice', 'Johnson', 'alice.johnson@example.com', 40);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Bob', 'Williams', 'bob.williams@example.com', 35);

INSERT INTO employees (first_name, last_name, email, age)
VALUES ('Charlie', 'Brown', 'charlie.brown@example.com', 28);

```
