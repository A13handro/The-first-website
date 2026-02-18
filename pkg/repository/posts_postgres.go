package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	todo "the-first-website"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
)

type PostPostgres struct {
	db          *sqlx.DB
	MinIOClient *minio.Client
}

func NewPostPostgres(db *sqlx.DB) *PostPostgres {
	return &PostPostgres{db: db}
}

func (r *PostPostgres) CreatePost(RefreshToken string, post todo.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var Role string
	var UserID uuid.UUID
	query := fmt.Sprintf("SELECT role, user_id FROM %s WHERE refresh_token = $1", usersTable)
	row := tx.QueryRow(query, RefreshToken)
	if err := row.Scan(&Role, &UserID); err != nil {
		tx.Rollback()
		return err
	}
	if Role != "Author" {
		tx.Rollback()
		return errors.New("Нет прав")
	}

	queryt := fmt.Sprintf("SELECT author_id FROM %s WHERE idempotency_key = $1", articlesTable)
	rot := tx.QueryRow(queryt, post.Idempotency_key)
	if err := rot.Scan(&UserID); err == nil {
		tx.Rollback()
		return errors.New("Ключ идемпотентности уже использован")
	}

	querys := fmt.Sprintf("INSERT INTO %s (title, content, created_at, updated_at, author_id, idempotency_key, status) VALUES ($1, $2, $3, $4, $5, $6, $7)", articlesTable)
	tx.QueryRow(querys, post.Title, post.Content, time.Now(), time.Now(), UserID, post.Idempotency_key, "Draft")
	fmt.Println("4")
	return tx.Commit()
}

func (r *PostPostgres) EditPost(RefreshToken string, post todo.Post, PostId string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var Role string
	var UserID uuid.UUID
	query := fmt.Sprintf("SELECT role, user_id FROM %s WHERE refresh_token = $1", usersTable)
	row := tx.QueryRow(query, RefreshToken)
	if err := row.Scan(&Role, &UserID); err != nil {
		tx.Rollback()
		return err
	}
	if Role != "Author" {
		return errors.New("Нет прав")
	}

	var post_id uuid.UUID
	table, err := tx.Query(fmt.Sprintf("SELECT post_id FROM %s WHERE author_id = $1", articlesTable), UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&post_id)
		if post_id.String() == PostId {
			eq = true
		}
	}
	table.Close()
	if err != nil || eq != true {
		tx.Rollback()
		return errors.New("Пост не найден")
	}

	_, err = tx.Exec(fmt.Sprintf("UPDATE %s SET title = $1, content = $2, updated_at = $3 WHERE post_id = $4 AND author_id = $5", articlesTable), post.Title, post.Content, time.Now(), PostId, UserID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostPostgres) PublishPost(RefreshToken string, post todo.Post, PostId string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var Role string
	var UserID uuid.UUID
	query := fmt.Sprintf("SELECT role, user_id FROM %s WHERE refresh_token = $1", usersTable)
	row := tx.QueryRow(query, RefreshToken)
	if err := row.Scan(&Role, &UserID); err != nil {
		tx.Rollback()
		return err
	}
	if Role != "Author" {
		return errors.New("Нет прав")
	}

	var post_id uuid.UUID
	table, err := tx.Query(fmt.Sprintf("SELECT post_id FROM %s WHERE author_id = $1", articlesTable), UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&post_id)
		if post_id.String() == PostId {
			eq = true
		}
	}
	table.Close()
	if err != nil || eq != true {
		tx.Rollback()
		return errors.New("Пост не найден")
	}

	if post.Status == "Published" {
		_, err = tx.Exec(fmt.Sprintf("UPDATE %s SET status = $1 WHERE post_id = $2", articlesTable), post.Status, PostId)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		tx.Rollback()
		return errors.New("Неверное значение статуса")
	}

	return tx.Commit()
}

func (r *PostPostgres) ViewingPosts(RefreshToken string) ([]byte, error) {

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	var Role string
	var UserID uuid.UUID
	query := fmt.Sprintf("SELECT role, user_id FROM %s WHERE refresh_token = $1", usersTable)
	row := tx.QueryRow(query, RefreshToken)
	if err := row.Scan(&Role, &UserID); err != nil {
		tx.Rollback()
		return nil, err
	}

	type Article struct {
		Title     string
		Content   string
		Createdat time.Time
		Updatedat time.Time
		Status    string
	}

	var Art = []Article{}
	switch Role {
	case "Reader":
		table, err := tx.Query(fmt.Sprintf("SELECT title, content, created_at, updated_at, post_id FROM %s WHERE status = 'Published'", articlesTable))
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		for table.Next() {
			var post Article
			var PostId uuid.UUID
			table.Scan(&post.Title, &post.Content, &post.Createdat, &post.Updatedat, &PostId)
			post.Status = "Published"
			Art = append(Art, post)
		}
		table.Close()

	case "Author":

		table, err := tx.Query(fmt.Sprintf("SELECT title, content, created_at, updated_at, status, post_id FROM %s WHERE author_id=$1", articlesTable), UserID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		for table.Next() {
			var post Article
			var Postik uuid.UUID
			table.Scan(&post.Title, &post.Content, &post.Createdat, &post.Updatedat, &post.Status, &Postik)
			Art = append(Art, post)
		}
		table.Close()

	}
	jsonData, err := json.MarshalIndent(Art, "", "    ")
	return jsonData, tx.Commit()
}
