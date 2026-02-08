package pst

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
)

// @Summary Создание поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Posts
// @Router /api/posts [post]
func Posts(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	Refreshtoken := cookie.Value

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	defer db.Close()

	var Role string
	var UserID uuid.UUID
	err = db.QueryRow("SELECT role, userid FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	if Role == "Reader" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не является автором",
		})
		w.Write(jsonData)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		fmt.Println("Не вся форма заполнена")
		return
	}
	idempotencykey := title + content + UserID.String()
	err = db.QueryRow("SELECT authorid FROM articles WHERE idempotencykey = $1", idempotencykey).Scan(&UserID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ключ идемпотентности уже использован",
		})
		w.Write(jsonData)
		return
	}

	_, err = db.Exec(
		"INSERT INTO articles (title, content, createdat, updatedat, authorid, idempotencykey, status) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		title, content, time.Now(), time.Now(), UserID, idempotencykey, "Draft",
	)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonData, _ := json.Marshal(map[string]string{
		"Message":        "Пост успешно создан",
		"title":          title,
		"content":        content,
		"createdat":      time.Now().String(),
		"updatedat":      time.Now().String(),
		"authorid":       UserID.String(),
		"idempotencykey": idempotencykey,
		"status":         "Draft",
	})
	w.Write(jsonData)
}

// @Summary Добавление картинки к посту
// @Description Требует аутентификации(Роль Автор).
// @Tags Images
// @Router /api/posts/{postId}/images [post]
func Images(w http.ResponseWriter, r *http.Request) {
	PostId := mux.Vars(r)["postId"]
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}
	Refreshtoken := cookie.Value

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}
	defer db.Close()

	var UserID uuid.UUID
	var Role string
	err = db.QueryRow("SELECT role, userid FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	if Role == "Reader" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Доступ запрещен",
		})
		w.Write(jsonData)
		return
	}
	var postid uuid.UUID
	table, err := db.Query("SELECT postid FROM articles WHERE authorid = $1", UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&postid)
		if postid.String() == PostId {
			eq = true
		}
	}
	if err != nil || eq != true {
		fmt.Println(postid)
		fmt.Println(PostId)
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пост не найден",
		})
		w.Write(jsonData)
		return
	}

	// Получаем файл
	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Ошибка получения файла: ", err)
	}
	defer file.Close()

	objectName := uuid.New()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	minioClient := ServerMini(w, r)

	//Сохраняем картинку в бд
	insert, err := db.Query("INSERT INTO pictures (imageid, postid, createdat) VALUES ($1, $2, $3)", objectName, PostId, time.Now())
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}
	insert.Close()

	//Выгружаем файл в minio
	_, err = minioClient.PutObject(context.Background(), "pictures", objectName.String(), file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		fmt.Println("Ошибка загрузки в MinIO: ", err)
	}

	fmt.Printf("Картинка %s загружена в бакет pictures\n", objectName)
	w.WriteHeader(http.StatusCreated)
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Картинка успешно добавлена",
	})
	w.Write(jsonData)
}

// @Summary Редактирование поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Edit
// @Router /api/posts/{postId} [put]
func Edit(w http.ResponseWriter, r *http.Request) {
	PostId := mux.Vars(r)["postId"]
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		fmt.Println("Ошибка: ", err)
		return
	}
	Refreshtoken := cookie.Value

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	var Role string
	err = db.QueryRow("SELECT role, userid FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	if Role == "Reader" || Role != "Author" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Доступ запрещен",
		})
		w.Write(jsonData)
		return
	}

	var postid uuid.UUID
	table, err := db.Query("SELECT postid FROM articles WHERE authorid = $1", UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&postid)
		if postid.String() == PostId {
			eq = true
		}
	}
	if err != nil || eq != true {
		fmt.Println(postid)
		fmt.Println(PostId)
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пост не найден",
		})
		w.Write(jsonData)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		fmt.Println("Не вся форма заполнена")
		return
	}
	_, err = db.Exec("UPDATE articles SET title = $1, content = $2, updatedat = $3 WHERE postid = $4 AND authorid = $5", title, content, time.Now(), PostId, UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"Message":   "Пост успешно обновлен",
		"title":     title,
		"content":   content,
		"updatedat": time.Now().String(),
	})
	w.Write(jsonData)
}

// @Summary Удаление картинки из поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Delete
// @Router /api/posts/{postId}/images/{imageId} [delete]
func Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) //postId
	PostId := vars["postId"]
	ImageId := vars["imageId"]

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}
	Refreshtoken := cookie.Value

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}
	defer db.Close()

	//Проверяем роль
	var Role string
	err = db.QueryRow("SELECT role FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	if Role == "Reader" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Доступ запрещен",
		})
		w.Write(jsonData)
		return
	}
	var postid uuid.UUID
	err = db.QueryRow("SELECT postid FROM pictures WHERE imageid = $1", ImageId).Scan(&postid)
	if err != nil || PostId != postid.String() {
		w.WriteHeader(http.StatusNotFound)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Картинка или пост не найдены",
		})
		w.Write(jsonData)
		return
	}

	//Удаляем картинку из бд
	_, err = db.Exec("DELETE FROM pictures WHERE postid = $1 AND imageid = $2", PostId, ImageId)
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}

	minioClient := ServerMini(w, r)

	objectName, _ := uuid.Parse(ImageId)
	ctx := context.Background()
	bucketName := "pictures"

	// Удаляем объект из бакета
	err = minioClient.RemoveObject(ctx, bucketName, objectName.String(), minio.RemoveObjectOptions{})
	if err != nil {
		fmt.Printf("Ошибка удаления файла %s: %v", objectName, err)
	}

	fmt.Printf("Картинка %s удалена из бакета pictures\n", objectName)
	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Картинка успешно удалена",
	})
	w.Write(jsonData)
}

// @Summary Публикация поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Publish
// @Router /api/posts/{postId}/status [PATCH]
func Publish(w http.ResponseWriter, r *http.Request) {
	PostId := mux.Vars(r)["postId"]
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		fmt.Println("Ошибка: ", err)
		return
	}
	Refreshtoken := cookie.Value

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	var Role string
	err = db.QueryRow("SELECT role, userid FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	if Role == "Reader" || Role != "Author" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Доступ запрещен",
		})
		w.Write(jsonData)
		return
	}

	var postid uuid.UUID
	table, err := db.Query("SELECT postid FROM articles WHERE authorid = $1", UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&postid)
		if postid.String() == PostId {
			eq = true
		}
	}
	if err != nil || eq != true {
		fmt.Println(postid)
		fmt.Println(PostId)
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пост не найден",
		})
		w.Write(jsonData)
		return
	}

	var Status string
	err = db.QueryRow("SELECT status FROM articles WHERE postid = $1", PostId).Scan(&Status)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	if Status == "Draft" {
		_, err = db.Exec("UPDATE articles SET status = 'Published' WHERE postid = $1", PostId)
		if err != nil {
			fmt.Println("Ошибка: ", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Неверное значение статуса",
			"Status":  "Draft",
		})
		w.Write(jsonData)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Пост успешно опубликован",
		"Status":  "Published",
	})
	w.Write(jsonData)
}
