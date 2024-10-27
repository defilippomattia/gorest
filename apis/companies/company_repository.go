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
