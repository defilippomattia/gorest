package companies

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepository interface {
	GetByID(ctx context.Context, id int) (*Company, error)
	Create(ctx context.Context, company *Company) error
	GetAll(ctx context.Context) ([]Company, error)
}

type PgCompanyRepository struct {
	db *pgxpool.Pool
}

func NewPgCompanyRepository(db *pgxpool.Pool) *PgCompanyRepository {
	return &PgCompanyRepository{db: db}
}

func (r *PgCompanyRepository) Create(ctx context.Context, company *Company) error {
	args := pgx.NamedArgs{
		"name":        company.Name,
		"yearFounded": company.YearFounded,
	}
	query := "INSERT INTO companies (name, year_founded) VALUES (@name, $yearFounded) RETURNING id"
	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (r *PgCompanyRepository) GetByID(ctx context.Context, id int) (*Company, error) {
	var company Company
	query := "SELECT id, name, year_founded FROM companies WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&company.ID, &company.Name, &company.YearFounded)
	if err != nil {
		return nil, fmt.Errorf("could not find company with id %d: %w", id, err)
	}
	return &company, nil
}

func (r *PgCompanyRepository) GetAll(ctx context.Context) ([]Company, error) {
	var companies []Company
	query := "SELECT id, name, year_founded FROM companies"
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve companies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var company Company
		if err := rows.Scan(&company.ID, &company.Name, &company.YearFounded); err != nil {
			return nil, fmt.Errorf("could not scan company row: %w", err)
		}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return companies, nil
}
