package pst

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не авторизован",
		})
		w.Write(jsonData)
		return
	}
	Refreshtoken := cookie.Value

	err = godotenv.Load()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ошибка .env",
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	defer db.Close()

	var Role string
	var UserID uuid.UUID
	err = db.QueryRow("SELECT role, user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
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
	idempotencykey := r.FormValue("idempotencykey")

	if title == "" || content == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Не вся форма заполнена",
		})
		w.Write(jsonData)
		return
	}
	err = db.QueryRow("SELECT author_id FROM articles WHERE idempotency_key = $1", idempotencykey).Scan(&UserID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ключ идемпотентности уже использован",
		})
		w.Write(jsonData)
		return
	}

	_, err = db.Exec(
		"INSERT INTO articles (title, content, created_at, updated_at, author_id, idempotency_key, status) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		title, content, time.Now(), time.Now(), UserID, idempotencykey, "Draft",
	)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonData, _ := json.Marshal(map[string]string{
		"Message":         "Пост успешно создан",
		"title":           title,
		"content":         content,
		"created_at":      time.Now().String(),
		"updated_at":      time.Now().String(),
		"author_id":       UserID.String(),
		"idempotency_key": idempotencykey,
		"status":          "Draft",
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
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не авторизован",
		})
		w.Write(jsonData)
	}
	Refreshtoken := cookie.Value

	err = godotenv.Load()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error":   err.Error(),
			"Message": "Ошибка загрузки .env",
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
	}
	defer db.Close()

	var UserID uuid.UUID
	var Role string
	err = db.QueryRow("SELECT role, user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
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
	table, err := db.Query("SELECT post_id FROM articles WHERE author_id = $1", UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&postid)
		if postid.String() == PostId {
			eq = true
		}
	}
	table.Close()
	if err != nil || eq != true {
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
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Картинка не получена",
		})
		w.Write(jsonData)
		return
	}
	defer file.Close()

	objectName := uuid.New()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	minioClient := ServerMini(w, r)

	//Выгружаем файл в minio
	_, err = minioClient.PutObject(context.Background(), "pictures", objectName.String(), file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error":   err.Error(),
			"Message": "Ошибка загрузки в MinIO",
		})
		w.Write(jsonData)
		return
	}

	image_url, err := minioClient.PresignedGetObject(
		context.Background(),
		"pictures",
		objectName.String(),
		604800*time.Second,
		nil,
	)

	//Сохраняем картинку в бд
	insert, err := db.Query("INSERT INTO pictures (image_id, post_id, created_at, image_url) VALUES ($1, $2, $3, $4)", objectName, PostId, time.Now(), image_url.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	insert.Close()

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
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не авторизован",
		})
		w.Write(jsonData)
		return
	}
	Refreshtoken := cookie.Value

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ошибка загрузки .env",
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	var Role string
	err = db.QueryRow("SELECT role, user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	if Role != "Author" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Доступ запрещен",
		})
		w.Write(jsonData)
		return
	}

	var postid uuid.UUID
	table, err := db.Query("SELECT post_id FROM articles WHERE author_id = $1", UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&postid)
		if postid.String() == PostId {
			eq = true
		}
	}
	table.Close()
	if err != nil || eq != true {
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
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Не вся форма заполнена",
		})
		w.Write(jsonData)
		return
	}
	_, err = db.Exec("UPDATE articles SET title = $1, content = $2, updated_at = $3 WHERE post_id = $4 AND author_id = $5", title, content, time.Now(), PostId, UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"Message":    "Пост успешно обновлен",
		"title":      title,
		"content":    content,
		"updated_at": time.Now().String(),
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
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не авторизован",
		})
		w.Write(jsonData)
	}
	Refreshtoken := cookie.Value

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ошибка загрузки .env",
			"Error":   err.Error(),
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
	}
	defer db.Close()

	//Проверяем роль
	var Role string
	err = db.QueryRow("SELECT role FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
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
	err = db.QueryRow("SELECT post_id FROM pictures WHERE image_id = $1", ImageId).Scan(&postid)
	if err != nil || PostId != postid.String() {
		w.WriteHeader(http.StatusNotFound)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Картинка или пост не найдены",
		})
		w.Write(jsonData)
		return
	}

	//Удаляем картинку из бд
	_, err = db.Exec("DELETE FROM pictures WHERE post_id = $1 AND image_id = $2", PostId, ImageId)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
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
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не авторизован",
		})
		w.Write(jsonData)
		return
	}
	Refreshtoken := cookie.Value

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ошибка загрузки .env",
			"Error":   err.Error(),
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	var Role string
	err = db.QueryRow("SELECT role, user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	if Role != "Author" {
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Доступ запрещен",
		})
		w.Write(jsonData)
		return
	}

	var postid uuid.UUID
	table, err := db.Query("SELECT post_id FROM articles WHERE author_id = $1", UserID)
	var eq bool = false
	for table.Next() {
		table.Scan(&postid)
		if postid.String() == PostId {
			eq = true
		}
	}
	table.Close()
	if err != nil || eq != true {
		w.WriteHeader(http.StatusNotFound)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пост не найден",
		})
		w.Write(jsonData)
		return
	}
	Status := r.FormValue("status")
	if Status == "Published" {
		_, err = db.Exec("UPDATE articles SET status = $1 WHERE post_id = $2", Status, PostId)
		if err != nil {
			fmt.Println("Ошибка: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonData, _ := json.Marshal(map[string]string{
				"Error": err.Error(),
			})
			w.Write(jsonData)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Неверное значение статуса",
			"Status":  Status,
		})
		w.Write(jsonData)
		return
	}
	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Пост успешно опубликован",
		"Status":  Status,
	})
	w.Write(jsonData)
}

// @Summary Просмотр постов
// @Description Требует аутентификации.
// @Tags Viewing
// @Router /api/posts [Get]
func Viewing(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Пользователь не авторизован",
		})
		w.Write(jsonData)
		return
	}
	Refreshtoken := cookie.Value

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ошибка загрузки .env",
			"Error":   err.Error(),
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	defer db.Close()

	var Role string
	var UserID uuid.UUID
	err = db.QueryRow("SELECT role, user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&Role, &UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}

	type Article struct {
		Title     string
		Content   string
		Createdat time.Time
		Updatedat time.Time
		Status    string
		Number    int
	}

	var Art = []Article{}
	switch Role {
	case "Reader":
		table, err := db.Query("SELECT title, content, created_at, updated_at, post_id FROM articles WHERE status = 'Published'")
		if err != nil {
			fmt.Println("Ошибка: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonData, _ := json.Marshal(map[string]string{
				"Error": err.Error(),
			})
			w.Write(jsonData)
			return
		}
		for table.Next() {
			var post Article
			var PostId uuid.UUID
			table.Scan(&post.Title, &post.Content, &post.Createdat, &post.Updatedat, &PostId)

			pic, err := db.Query("SELECT image_id FROM pictures WHERE post_id = $1", PostId)
			if err != nil {
				fmt.Println("Ошибка: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonData, _ := json.Marshal(map[string]string{
					"Error": err.Error(),
				})
				w.Write(jsonData)
				return
			}
			for pic.Next() { //Подсчет количества картинок в посте
				post.Number++
			}
			Art = append(Art, post)
		}
		table.Close()

	case "Author":

		table, err := db.Query("SELECT title, content, created_at, updated_at, status, post_id FROM articles WHERE author_id=$1", UserID)
		if err != nil {
			fmt.Println("Ошибка: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonData, _ := json.Marshal(map[string]string{
				"Error": err.Error(),
			})
			w.Write(jsonData)
			return
		}

		for table.Next() {
			var post Article
			var PostId uuid.UUID
			table.Scan(&post.Title, &post.Content, &post.Createdat, &post.Updatedat, &post.Status, &PostId)
			pic, err := db.Query("SELECT image_id FROM pictures WHERE post_id = $1", PostId)
			if err != nil {
				fmt.Println("Ошибка: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonData, _ := json.Marshal(map[string]string{
					"Error": err.Error(),
				})
				w.Write(jsonData)
				return
			}
			for pic.Next() { //Подсчет количества картинок в посте
				post.Number++
			}
			Art = append(Art, post)
		}
		table.Close()
	}

	w.WriteHeader(http.StatusOK)
	jsonData, err := json.MarshalIndent(Art, "", "    ")
	w.Write(jsonData)
}
