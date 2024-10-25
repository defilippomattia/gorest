package books

import (
	"encoding/json"
	"net/http"

	"github.com/defilippomattia/gorest/apis"
)

func GetBooks(sd *apis.ServerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		books, err := fetchBooks(sd.Conn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := &BooksOutput{Books: books}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
