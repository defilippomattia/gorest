package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func ConnectToDatabase(connUrl string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), connUrl)
	if err != nil {
		log.Error().Err(err).Msg("unable to create connection pool")
		return nil, nil
	}
	return dbpool, nil
}
