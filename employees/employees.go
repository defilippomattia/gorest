package employees

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Employee struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

type EmployeeInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
}

type EmployeesInput struct {
	Session http.Cookie `cookie:"session_token"` // Use the correct cookie name here
}

type EmployeesOutput struct {
	Body struct {
		Employees []Employee `json:"employees"`
	}
}

type EmployeeOutput struct {
	Body Employee `json:"body"`
}

func GetEmployees(conn *pgxpool.Pool) func(ctx context.Context, input *EmployeesInput) (*EmployeesOutput, error) {
	return func(ctx context.Context, input *EmployeesInput) (*EmployeesOutput, error) {
		fmt.Printf("session token: %v\n", input.Session.Value)
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
	}
}

func GetEmployeeById(conn *pgxpool.Pool) func(ctx context.Context, input *struct {
	ID int `path:"id"`
}) (*EmployeeOutput, error) {
	return func(ctx context.Context, input *struct {
		ID int `path:"id"`
	}) (*EmployeeOutput, error) {
		row := conn.QueryRow(context.Background(), "SELECT id, first_name, last_name, email, age, created_at FROM employees WHERE id = $1", input.ID)
		fmt.Printf("id: %v\n", input.ID)
		fmt.Printf("row: %v\n", row)

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

		fmt.Printf("employee: %v\n", employee)

		resp := &EmployeeOutput{
			Body: employee,
		}
		return resp, nil
	}
}

func CreateEmployee(conn *pgxpool.Pool) func(ctx context.Context, input *EmployeeInput) (*EmployeeOutput, error) {
	return func(ctx context.Context, input *EmployeeInput) (*EmployeeOutput, error) {
		var employeeID int
		err := conn.QueryRow(context.Background(),
			"INSERT INTO employees (first_name, last_name, email, age) VALUES ($1, $2, $3, $4) RETURNING id",
			input.FirstName, input.LastName, input.Email, input.Age).Scan(&employeeID)

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
	}
}
