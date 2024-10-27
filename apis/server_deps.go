package apis

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerDeps struct {
	Conn      *pgxpool.Pool
	Something string
}
