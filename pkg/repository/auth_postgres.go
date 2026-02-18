package repository

import (
	"errors"
	"fmt"
	todo "the-first-website"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(RefreshToken string, user todo.User) error {
	var id uuid.UUID

	if user.IsEmail == false {
		return errors.New("Неверный формат email")
	}

	table := fmt.Sprintf("SELECT user_id FROM %s WHERE email = $1", usersTable)
	rot := r.db.QueryRow(table, user.Email)
	if err := rot.Scan(&id); err == nil {
		return errors.New("Email уже существует")
	}
	refresh_token_expiry_time := time.Now().Add(7 * 24 * time.Hour)
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash, role, refresh_token, refresh_token_expiry_time) values ($1, $2, $3, $4, $5) RETURNING user_id", usersTable)
	row := r.db.QueryRow(query, user.Email, user.Password, user.Role, RefreshToken, refresh_token_expiry_time)
	err := row.Scan(&id)
	return err
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

func (r *AuthPostgres) UpdateUser(Refreshtoken string, UserID uuid.UUID) error {
	_, err := r.db.Exec(fmt.Sprintf("UPDATE %s SET refresh_token = $1, refresh_token_expiry_time = CURRENT_TIMESTAMP + INTERVAL '7 days' WHERE user_id = $2", usersTable), Refreshtoken, UserID)
	return err
}
