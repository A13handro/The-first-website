package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Article struct {
	Id                     uint16
	Title, Anons, FullText string
}

var posts = []Article{}
var showPosts = Article{}

var ThisToken string
var ThiseID int
var Message string
var secretKey = []byte("my_secret_key")

func generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		var authToken string
		erro := db.QueryRow("SELECT auth_token FROM users WHERE id = $1", ThiseID).Scan(&authToken)
		if erro != nil {
			Message = "Токен не получен!"
			http.Redirect(w, r, "/authorization", http.StatusSeeOther)
		}
		if authToken == ThisToken {
			next(w, r)
		} else {
			fmt.Println("FFFFFFFFFf")
		}
	}
}

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
	t.ExecuteTemplate(w, "authorization", Message)
	Message = ""
}

func auth(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	if login == "" || password == "" {
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/authorization", http.StatusSeeOther)
		return
	}
	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	insert, err := db.Query("SELECT id FROM users WHERE login = $1 AND hashed_password = $2", login, hashedPass)
	if err != nil {
		fmt.Println("НЕТ!")
		panic(err)
	}
	if insert.Next() {
		var id int
		err := insert.Scan(&id)
		if err != nil {
			panic(err)
		}
		ThiseID = id
		ThisToken, _ = generateToken(id)
		_, err = db.Exec("UPDATE users SET auth_token = $1 WHERE id = $2", ThisToken, id)
		if err != nil {
			panic(err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		Message = "Неверный логин или пароль!"
		http.Redirect(w, r, "/authorization", http.StatusSeeOther)
	}
	insert.Close()
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

	t.ExecuteTemplate(w, "create", Message)
	Message = ""
}

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	if title == "" || anons == "" || full_text == "" {
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/create", http.StatusSeeOther)
		return
	}
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

	rtr.HandleFunc("/", authMiddleware(index)).Methods("GET")
	rtr.HandleFunc("/create", authMiddleware(create)).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", authMiddleware(show_post)).Methods("GET")
	rtr.HandleFunc("/delete_post/{id:[0-9]+}", authMiddleware(delete_post)).Methods("GET")
	rtr.HandleFunc("/authorization", authorization).Methods("GET")
	rtr.HandleFunc("/registration", registration).Methods("GET")
	rtr.HandleFunc("/auth", auth).Methods("POST")

	http.Handle("/", rtr)

	http.ListenAndServe(":8080", nil)

}

func main() {
	handlfunc()
}
