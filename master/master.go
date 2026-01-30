package mst

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	tkns "the-first-website/tokens"

	_ "github.com/lib/pq"
)

// @Summary Главная страница
// @Tags Master
// @Router /api [get]
func Master(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/header.html", "templates/index.html", "templates/footer.html")
	tkns.CheckErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err1 := sql.Open("postgres", connStr)
	tkns.CheckErr(err1)
	defer db.Close()

	table, err2 := db.Query("SELECT title, content, images, createdat, updatedat, authorid, postid FROM articles")
	tkns.CheckErr(err2)

	var role string
	var Art = []tkns.Article{}
	err3 := db.QueryRow("SELECT role FROM users WHERE userid = $1", tkns.UserID).Scan(&role)
	tkns.CheckErr(err3)
	for table.Next() { //Заполняем Art
		var post tkns.Article
		var cr, up time.Time
		err = table.Scan(&post.Title, &post.Content, &post.Images, &cr, &up, &post.Authorid, &post.PostId)
		post.Rollle = role
		err4 := db.QueryRow("SELECT name, surname FROM users WHERE userid = $1", post.Authorid).Scan(&post.Name, &post.Surname)
		tkns.CheckErr(err4)
		post.Createdat = cr.Format(time.Stamp)
		post.Updatedat = up.Format(time.Stamp)
		tkns.CheckErr(err)
		Art = append(Art, post)
	}
	data := struct { //Создаем структуру для передачи данных в html
		Posts []tkns.Article
		Role  string
		Mes   string
	}{
		Posts: Art,
		Role:  role,
		Mes:   tkns.Message,
	}
	t.ExecuteTemplate(w, "index", data)
	tkns.Message = ""
}
