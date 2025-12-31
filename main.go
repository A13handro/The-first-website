package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Article struct {
	Id             uint16
	Title, Сontent string
	Role           bool
	Pictures       *string
}

var postss = []Article{}
var showPosts = Article{}

var AccessToken string
var RefreshToken string

var ThiseID int
var Message string
var secretKey = []byte("my_secret_key")

func checkErr(err error) { //Проверка на ошибку
	if err != nil {
		Message = "Ошибка"
		fmt.Println("Ошибка:")
		panic(err)
	}
}

func generateToken(userID int, num int) (string, error) { //Создание токена
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(time.Hour * time.Duration(num)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
func parseToken(tokenString string) (*jwt.Token, error) { //Проверка токена
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}
func authMiddleware(next http.HandlerFunc) http.HandlerFunc { //Проверка токена
	return func(w http.ResponseWriter, r *http.Request) {
		connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		checkErr(err)
		defer db.Close()

		var authTokenA string
		var authTokenR string
		table, err := db.Query("SELECT accesstoken FROM users WHERE id = $1", ThiseID)
		checkErr(err)
		if table.Next() {
			table.Scan(&authTokenA)
		}
		if AccessToken == "" || AccessToken != authTokenA {
			Message = "Пожалуйста, авторизируйтесь!"
			http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
			return
		}
		tokenA, _ := parseToken(authTokenA)
		if !tokenA.Valid {
			err1 := db.QueryRow("SELECT refreshtoken FROM users WHERE id = $1", ThiseID).Scan(&authTokenR)
			checkErr(err1)
			tokenR, _ := parseToken(authTokenR)
			if tokenR.Valid {
				AccessToken, _ = generateToken(ThiseID, 15)
				_, err = db.Exec("UPDATE users SET accesstoken = $1 WHERE id = $2", AccessToken, ThiseID)
				checkErr(err)
			} else {
				Message = "Время сеанса истекло. Пожалуйста, авторезируйтесь повторно."
				http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
				return
			}
		}
		next(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) { //Главная страница
	t, err := template.ParseFiles("templates/header.html", "templates/index.html", "templates/footer.html")
	checkErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err1 := sql.Open("postgres", connStr)
	checkErr(err1)
	defer db.Close()

	table, err2 := db.Query("SELECT * FROM articles")
	checkErr(err2)
	postss = []Article{}
	for table.Next() {
		var post Article
		err = table.Scan(&post.Id, &post.Title, &post.Сontent, &post.Pictures)
		checkErr(err)
		err3 := db.QueryRow("SELECT role FROM users WHERE id = $1", ThiseID).Scan(&post.Role)
		checkErr(err3)
		postss = append(postss, post)
	}
	t.ExecuteTemplate(w, "index", postss)
}

func login(w http.ResponseWriter, r *http.Request) { //Авторизация
	t, err := template.ParseFiles("templates/header.html", "templates/login.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "login", Message)
	Message = ""
}

func log(w http.ResponseWriter, r *http.Request) { //обработка Авторизации
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
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

	insert, err := db.Query("SELECT id FROM users WHERE email = $1 AND passwordhash = $2", email, hashedPass)
	if err != nil {
		panic(err)
	}
	if insert.Next() {
		var id int
		err := insert.Scan(&id)
		if err != nil {
			panic(err)
		}
		ThiseID = id
		AccessToken, _ = generateToken(id, 2)
		RefreshToken, _ = generateToken(id, 168)
		_, err = db.Exec("UPDATE users SET accesstoken = $1, refreshtoken = $2 WHERE id = $3", AccessToken, RefreshToken, id)
		if err != nil {
			panic(err)
		}
		http.Redirect(w, r, "/api", http.StatusSeeOther)
	} else {
		Message = "Неверный логин или пароль!"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
	}
}

func register(w http.ResponseWriter, r *http.Request) { //Регистрация
	t, err := template.ParseFiles("templates/register.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "register", Message)
	Message = ""
}

func reg(w http.ResponseWriter, r *http.Request) { //обработка Регистрации
	email := r.FormValue("email")
	em := govalidator.IsEmail(strings.TrimSpace(email))
	password := r.FormValue("password")
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	roleStr := r.FormValue("role")
	if email == "" || password == "" || name == "" || surname == "" {
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	} else if em == false {
		Message = "Формат логина не соответсвует!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	}
	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	var author bool
	if roleStr == "true" {
		author = true
	} else {
		author = false
	}
	fmt.Println(author)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var idd string
	erro := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&idd)
	if erro == nil {
		Message = "Такой email уже существует!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	}
	insert, err := db.Query("INSERT INTO users (email, passwordhash, name, surname, role) "+
		"VALUES ($1, $2, $3, $4, $5)", email, hashedPass, name, surname, author)
	if err != nil {
		panic(err)
	}
	insert.Close()
	http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
}

func posts(w http.ResponseWriter, r *http.Request) { //Создание поста
	t, err := template.ParseFiles("templates/header.html", "templates/posts.html", "templates/footer.html")
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "posts", Message)
	Message = ""
}

func pos(w http.ResponseWriter, r *http.Request) { //обработка Создания или изменения поста
	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	pictures := r.FormValue("pictures")
	title := r.FormValue("title")
	content := r.FormValue("content")
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	adres := r.FormValue("a")

	if title == "" || content == "" {
		if adres != "/api/posts" {
			http.Redirect(w, r, fmt.Sprintf("/delete_post/%d", id), http.StatusSeeOther)
			return
		}
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/posts", http.StatusSeeOther)
		return
	}
	err = db.Ping()
	checkErr(err)

	// if pictures == "" {
	// 	pictures = "s"
	// }

	ins, err1 := db.Query("SELECT * FROM articles WHERE id = $1", id)
	checkErr(err1)
	if ins.Next() {
		_, err = db.Exec("UPDATE articles SET title = $1, content = $2, pictures = $3 WHERE id = $4", title, content, pictures, id)
	} else {
		insert, err := db.Query("INSERT INTO articles (title, content, pictures) VALUES ($1, $2, $3)", title, content, pictures)
		checkErr(err)
		insert.Close()
	}

	http.Redirect(w, r, "/api", http.StatusSeeOther)
}

func edit_post(w http.ResponseWriter, r *http.Request) {
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
		err = insert.Scan(&post.Id, &post.Title, &post.Сontent, &post.Pictures)
		checkErr(err)
		showPosts = post
	}

	t.ExecuteTemplate(w, "show", showPosts)
}

func delete_post(w http.ResponseWriter, r *http.Request) { //Удаление поста
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

	http.Redirect(w, r, "/api", http.StatusSeeOther)
}

func handlfunc() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/api", authMiddleware(index)).Methods("GET")       //Главная страница
	rtr.HandleFunc("/api/posts", authMiddleware(posts)).Methods("GET") //Создание поста
	rtr.HandleFunc("/api/pos", pos).Methods("POST")                    //обработка Создания поста
	rtr.HandleFunc("/post/{id:[0-9]+}", authMiddleware(edit_post)).Methods("GET")
	rtr.HandleFunc("/delete_post/{id:[0-9]+}", authMiddleware(delete_post)).Methods("GET") //Удаление поста
	rtr.HandleFunc("/api/auth/login", login).Methods("GET")                                //Авторизация
	rtr.HandleFunc("/api/auth/log", log).Methods("POST")                                   //обработка Авторизации
	rtr.HandleFunc("/api/auth/register", register).Methods("GET")                          //Регистрация
	rtr.HandleFunc("/api/auth/reg", reg).Methods("POST")                                   //обработка Регистрации

	http.Handle("/", rtr)

	http.ListenAndServe(":8080", nil)

}

func main() {
	handlfunc()
}
