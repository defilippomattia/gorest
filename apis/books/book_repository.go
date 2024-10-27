package books

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func fetchBooks(conn *pgxpool.Pool) ([]Book, error) {
	books := []Book{}
	log.Info().Msg("fetchBooks called")
	query := "SELECT id, title, author FROM books"
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author); err != nil {
			log.Error().Err(err).Msg("error scanning row")
			return nil, err
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error scanning rows")
		return nil, err
	}
	log.Info().
		Str("query", query).
		Msg("query executed successfully")

	log.Info().Msg("fetchBooks completed")
	return books, nil
}
