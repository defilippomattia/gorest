package users

import (
	"context"
	"errors"

	"github.com/defilippomattia/gorest/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type UserRepository interface {
	Register(ctx context.Context, usRegReq *UserRegistrationRequest) (int, error)
}

type PgUserRepository struct {
	db *pgxpool.Pool
}

func NewPgUserRepository(db *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) Register(ctx context.Context, user *UserRegistrationRequest) (int, error) {

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		log.Error().Err(err).Msg("error hashing password")
		return -1, err
	}

	args := pgx.NamedArgs{
		"username": user.Username,
		"password": hashedPassword,
	}
	query := "INSERT INTO users (username, password) VALUES (@username, @password) RETURNING id"
	var userID int

	err = r.db.QueryRow(ctx, query, args).Scan(&userID)

	if err != nil {
		pgErr, isPgError := err.(*pgconn.PgError)
		if isPgError && pgErr.Code == "23505" {
			log.Error().Str("username", user.Username).Msg("username already exists")
			return -1, errors.New("username already exists")
		}
		log.Error().Err(err).Msg("error inserting new user")
		return -1, err
	}

	return userID, nil

}
