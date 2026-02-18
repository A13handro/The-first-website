package repository

import (
	"context"
	"errors"
	"fmt"
	todo "the-first-website"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
)

type PicturePostgres struct {
	db          *sqlx.DB
	MinIOClient *minio.Client
}

func NewPicturePostgres(db *sqlx.DB, minioClient *minio.Client) *PicturePostgres {
	return &PicturePostgres{
		db:          db,
		MinIOClient: minioClient}
}

func (r *PicturePostgres) AddImage(RefreshToken string, img todo.Image, PostId string) error {
	if r.MinIOClient == nil {
		return errors.New("MinIO-клиент не инициализирован")
	}
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
		return errors.New("Пост не найден")
	}

	objectName := uuid.New()
	contentType := img.FileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	//Выгружаем файл в minio
	_, err = r.MinIOClient.PutObject(context.Background(), "pictures", objectName.String(), img.File, img.FileHeader.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		tx.Rollback()
		return err
	}

	image_url, err := r.MinIOClient.PresignedGetObject(
		context.Background(),
		"pictures",
		objectName.String(),
		604800*time.Second,
		nil,
	)
	querys := fmt.Sprintf("INSERT INTO %s (image_id, post_id, created_at, image_url) VALUES ($1, $2, $3, $4) RETURNING post_id", picturesTable)
	ros := r.db.QueryRow(querys, objectName, PostId, time.Now(), image_url.String())
	err = ros.Scan(&PostId)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *PicturePostgres) DeleteImage(RefreshToken string, ImageId string, PostId string) error {
	if r.MinIOClient == nil {
		return errors.New("MinIO-клиент не инициализирован")
	}
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
	if err != nil {
		return err
	}
	var eq bool = false
	for table.Next() {
		err = table.Scan(&post_id)
		if err != nil {
			return err
		}
		if post_id.String() == PostId {
			eq = true
		}
	}
	table.Close()
	if err != nil || eq != true {
		return errors.New("Пост не найден")
	}

	var image_id uuid.UUID
	tablek, err := tx.Query(fmt.Sprintf("SELECT image_id FROM %s WHERE post_id = $1", picturesTable), PostId)
	if err != nil {
		return err
	}
	var eqq bool = false
	for tablek.Next() {
		err = tablek.Scan(&image_id)
		if err != nil {
			return err
		}
		if image_id.String() == ImageId {
			eqq = true
		}
	}
	tablek.Close()
	if err != nil || eqq != true {
		return errors.New("Картинка не найдена")
	}

	_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE post_id = $1 AND image_id = $2", picturesTable), PostId, ImageId)
	if err != nil {
		return err
	}

	objectName, _ := uuid.Parse(ImageId)
	ctx := context.Background()
	bucketName := "pictures"

	// Удаляем объект из бакета
	err = r.MinIOClient.RemoveObject(ctx, bucketName, objectName.String(), minio.RemoveObjectOptions{})
	if err != nil {
		fmt.Printf("Ошибка удаления файла %s: %v", objectName, err)
	}

	return tx.Commit()
}
