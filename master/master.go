package mst

import (
	"database/sql"
	"html/template"
	"net/http"
	tkns "the-first-website/tokens"

	_ "github.com/lib/pq"
)

type Article struct {
	Id             uint16
	Title, Content string
	Data           string
	Mes            string
	Rol            bool
}

func Master(w http.ResponseWriter, r *http.Request) { //Главная страница (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/index.html", "templates/footer.html")
	tkns.CheckErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err1 := sql.Open("postgres", connStr)
	tkns.CheckErr(err1)
	defer db.Close()

	table, err2 := db.Query("SELECT * FROM articles")
	tkns.CheckErr(err2)

	var role bool
	var Art = []tkns.Article{}
	err3 := db.QueryRow("SELECT role FROM users WHERE id = $1", tkns.ThiseID).Scan(&role)
	tkns.CheckErr(err3)
	for table.Next() { //Заполняем Art
		var post tkns.Article
		err = table.Scan(&post.Id, &post.Title, &post.Content, &post.Data)
		tkns.CheckErr(err)
		post.Rol = role
		Art = append(Art, post)
	}
	data := struct { //Создаем структуру для передачи данных в html
		Posts []tkns.Article
		Role  bool
		Mes   string
	}{
		Posts: Art,
		Role:  role,
		Mes:   tkns.Message,
	}
	t.ExecuteTemplate(w, "index", data)
	tkns.Message = ""
}
