package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	usersTable    = "users"
	articlesTable = "articles"
	picturesTable = "pictures"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s host=%s",
		cfg.Username, cfg.Password, cfg.Port, cfg.DBName, cfg.SSLMode, cfg.Host))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
