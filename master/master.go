package mst

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	err = db.QueryRow("SELECT role FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&Role)
	if err != nil {
		fmt.Println("Ошибка!:", err)
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
		table, err := db.Query("SELECT title, content, createdat, updatedat, postid FROM articles WHERE status = 'Published'")
		if err != nil {
			fmt.Println("Ошибка!:", err)
			return
		}
		for table.Next() {
			var post Article
			var PostId uuid.UUID
			table.Scan(&post.Title, &post.Content, &post.Createdat, &post.Updatedat, &PostId)

			pic, err := db.Query("SELECT imageid FROM pictures WHERE postid = $1", PostId)
			if err != nil {
				fmt.Println("Ошибка!:", err)
				return
			}
			for pic.Next() { //Подсчет количества картинок в посте
				post.Number++
			}
			Art = append(Art, post)
		}
		table.Close()

	case "Author":
		table, err := db.Query("SELECT title, content, createdat, updatedat, status, postid FROM articles")
		if err != nil {
			fmt.Println("Ошибка!:", err)
			return
		}

		for table.Next() {
			var post Article
			var PostId uuid.UUID
			table.Scan(&post.Title, &post.Content, &post.Createdat, &post.Updatedat, &post.Status, &PostId)
			pic, err := db.Query("SELECT imageid FROM pictures WHERE postid = $1", PostId)
			if err != nil {
				fmt.Println("Ошибка!:", err)
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
