package books

type BooksOutput struct {
	Books []Book `json:"books"`
}

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}
