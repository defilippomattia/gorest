package books

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func fetchBooks(conn *pgx.Conn) ([]Book, error) {
	books := []Book{}
	query := "SELECT id, title, author FROM books"

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
