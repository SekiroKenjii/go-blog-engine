package abstract

import "database/sql"

type IDatabase interface {
	Postgres() *sql.DB
}
