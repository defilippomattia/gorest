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

	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("failed to ping database check credentials and connectivity")
		dbpool.Close()
		return nil, err
	}

	return dbpool, nil
}
