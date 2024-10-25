package apis

import (
	"github.com/jackc/pgx/v5"
)

type ServerDeps struct {
	Conn      *pgx.Conn
	Something string
}
