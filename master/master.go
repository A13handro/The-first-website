package mst

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
