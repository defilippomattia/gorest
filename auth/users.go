package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

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

type LoginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Token    string `json:"token"`
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
	} `json:"body"`
}

func Register(conn *pgxpool.Pool) func(ctx context.Context, input *struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"body"`
}) (*UserOutput, error) {
	return func(ctx context.Context, input *struct {
		Body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"body"`
	}) (*UserOutput, error) {
		// var username string
		hashedPassword, err := HashPassword(input.Body.Password)
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
	}
}

func Login(conn *pgxpool.Pool) func(ctx context.Context, input *struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"body"`
}) (*LoginOutput, error) {
	return func(ctx context.Context, input *struct {
		Body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"body"`
	}) (*LoginOutput, error) {
		hashedPassword, err := HashPassword(input.Body.Password)
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

		token := GenerateSessionToken()
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
	}
}
