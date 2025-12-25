package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

type Article struct {
	Id                     uint16
	Title, Anons, FullText string
}

type Mes struct {
	Message string
}

var data = Mes{}
var posts = []Article{}
var showPosts = Article{}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/header.html", "templates/index.html", "templates/footer.html")
	if err != nil {
		panic(err)
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	table, err := db.Query("SELECT * FROM articles")
	if err != nil {
		panic(err)
	}
	posts = []Article{}
	for table.Next() {
		var post Article
		err = table.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}

	t.ExecuteTemplate(w, "index", posts)
}

func authorization(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/authorization.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, "authorization")
}

func registration(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/registration.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, "registration")
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/header.html", "templates/create.html", "templates/footer.html")
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "create", data)
	data = Mes{""}
}

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	if title == "" || anons == "" || full_text == "" {
		data = Mes{"Не все поля заполнены!"}
		http.Redirect(w, r, "/create", http.StatusSeeOther)
	} else {
		connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			panic(err)
		}

		insert, err := db.Query("INSERT INTO articles (title, anons, full_text) "+
			"VALUES ($1, $2, $3)", title, anons, full_text)
		if err != nil {
			panic(err)
		}
		insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func show_post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := template.ParseFiles("templates/header.html", "templates/show.html", "templates/footer.html")
	if err != nil {
		panic(err)
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	insert, err := db.Query("SELECT * FROM articles WHERE id = $1", vars["id"])
	if err != nil {
		panic(err)
	}

	showPosts = Article{}
	for insert.Next() {
		var post Article
		err = insert.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}

		showPosts = post
	}

	t.ExecuteTemplate(w, "show", showPosts)
}

func delete_post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	insert, err := db.Query("DELETE FROM articles WHERE id = $1", vars["id"])
	if err != nil {
		panic(err)
	}
	insert.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handlfunc() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", show_post).Methods("GET")
	rtr.HandleFunc("/delete_post/{id:[0-9]+}", delete_post).Methods("GET")
	rtr.HandleFunc("/authorization", authorization).Methods("GET")
	rtr.HandleFunc("/registration", registration).Methods("GET")

	http.Handle("/", rtr)

	http.ListenAndServe(":8080", nil)

}

func main() {
	handlfunc()
}
