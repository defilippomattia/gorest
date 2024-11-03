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
	Login(ctx context.Context, usLogReq *UserLoginRequest) (string, error)
	ValidateSessionToken(ctx context.Context, sessionToken string) (int, error)
}

type PgUserRepository struct {
	db *pgxpool.Pool
}

func NewPgUserRepository(db *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) Login(ctx context.Context, user *UserLoginRequest) (string, error) {
	sessionToken := auth.GenerateSessionToken()

	getUserArgs := pgx.NamedArgs{
		"username": user.Username,
	}

	getUserQuery := "SELECT id, username, password FROM users WHERE username = @username"
	//var passwordInDb string
	var userInDb User

	err := r.db.QueryRow(ctx, getUserQuery, getUserArgs).Scan(&userInDb.ID, &userInDb.Username, &userInDb.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error().Str("username", user.Username).Msg("username not found")
			return "", errors.New("username and password do not match")
		}
		log.Error().Err(err).Msg("error getting password from database")
		return "", err
	}

	match, err := auth.ComparePasswordAndHash(user.Password, userInDb.Password)
	if err != nil {
		log.Error().Err(err).Msg("error comparing password and hash")
		return "", err
	}

	if !match {
		log.Error().Str("username", user.Username).Msg("username and password do not match")
		return "", errors.New("username and password do not match")
	}

	//todo: check if session already exists for user and delete it maybe?

	insertSessionArgs := pgx.NamedArgs{
		"user_id": userInDb.ID,
		"token":   sessionToken,
	}

	insertSessionQuery := "INSERT INTO sessions (user_id, token) VALUES (@user_id, @token)"

	_, err = r.db.Exec(ctx, insertSessionQuery, insertSessionArgs)
	if err != nil {
		log.Error().Err(err).Msg("error inserting session into database")
		return "", err
	}

	log.Info().Str("username", user.Username).Msg("user logged in")

	return sessionToken, nil
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

func (r *PgUserRepository) ValidateSessionToken(ctx context.Context, sessionToken string) (int, error) {
	var userId int
	args := pgx.NamedArgs{
		"token": sessionToken,
	}

	query := "SELECT user_id FROM sessions WHERE token = @token"

	err := r.db.QueryRow(ctx, query, args).Scan(&userId)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error().Str("session_token", sessionToken).Msg("session token not found")
			return -1, errors.New("session token not found")
		}
		log.Error().Err(err).Msg("error getting user_id from session token")
		return -1, err
	}

	return userId, nil
}
