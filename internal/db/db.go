package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
)

var (
	dbInstance *Database
	dbOnce     sync.Once
)

type Database struct {
	pg *sql.DB
}

func DatabaseInstance() abstract.IDatabase {
	dbOnce.Do(func() {
		dbInstance = &Database{
			pg: newPostgres(),
		}
	})

	return dbInstance
}

func newPostgres() *sql.DB {
	pgConf := config.Instance().Postgres
	dsn := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=disable",
		pgConf.User,
		pgConf.Password,
		pgConf.Host,
		pgConf.Port,
		pgConf.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(pgConf.MaxIdleConns)
	db.SetMaxOpenConns(pgConf.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(pgConf.ConnMaxLifetime))

	return db
}

func (db *Database) Postgres() *sql.DB {
	return db.pg
}
