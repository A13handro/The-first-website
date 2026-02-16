package repository

import (
	"errors"
	"fmt"
	todo "the-first-website"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user todo.User) (uuid.UUID, error) {
	var id uuid.UUID

	if user.IsEmail == false {
		return id, errors.New("Неверный формат email")
	}

	table := fmt.Sprintf("SELECT user_id FROM %s WHERE email = $1", usersTable)
	rot := r.db.QueryRow(table, user.Email)
	if err := rot.Scan(&id); err == nil {
		return id, errors.New("Email уже существует")
	}
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, role) values ($1, $2, $3) RETURNING user_id", usersTable)
	row := r.db.QueryRow(query, user.Email, user.Password, user.Role)
	if err := row.Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

func (r *AuthPostgres) CheckUser(user todo.User) (uuid.UUID, string, error) {
	var pass string
	var UserID uuid.UUID
	query := fmt.Sprintf("SELECT password_hash, user_id FROM %s WHERE email = $1", usersTable)
	rot := r.db.QueryRow(query, user.Email)
	err := rot.Scan(&pass, &UserID)
	return UserID, pass, err
}

func (r *AuthPostgres) GetID(refresh_token string) (uuid.UUID, error) {
	var UserID uuid.UUID
	query := fmt.Sprintf("SELECT user_id FROM %s WHERE refresh_token = $1", usersTable)
	rot := r.db.QueryRow(query, refresh_token)
	err := rot.Scan(&UserID)
	return UserID, err
}
