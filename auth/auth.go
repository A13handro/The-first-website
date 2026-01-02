package auth

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
	tkns "the-first-website/tokens"

	"github.com/asaskevich/govalidator"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) { //Авторизация (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/login.html")
	tkns.CheckErr(err)

	tkns.AccessToken = "" // Выходим из аккаунта
	tkns.RefreshToken = ""

	t.ExecuteTemplate(w, "login", tkns.Message)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

func Log(w http.ResponseWriter, r *http.Request) { //обработка Авторизации (POST)
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" { //Проверка на заполнение формы
		tkns.Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	var storedHash string
	var id int
	err = db.QueryRow("SELECT passwordhash, id FROM users WHERE email = $1", email).Scan(&storedHash, &id) //Поиск
	if err != nil {
		tkns.Message = "Неверный логин"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) //Сравение пароля
	if err != nil {
		tkns.Message = "Неверный пароль"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}

	tkns.ThiseID = id                                  // Запоминаем ID пользователя
	tkns.AccessToken, _ = tkns.GenerateToken(id, 2)    // Создаем токен на 2 часа
	tkns.RefreshToken, _ = tkns.GenerateToken(id, 168) // Создаем токен на неделю
	_, err = db.Exec("UPDATE users SET accesstoken = $1, refreshtoken = $2 WHERE id = $3", tkns.AccessToken, tkns.RefreshToken, id)
	tkns.CheckErr(err) //Отправляем токен в бд
	http.Redirect(w, r, "/api", http.StatusSeeOther)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

func Register(w http.ResponseWriter, r *http.Request) { //Регистрация (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/register.html")
	tkns.CheckErr(err)

	tkns.AccessToken = "" // Выходим из аккаунта
	tkns.RefreshToken = ""

	t.ExecuteTemplate(w, "register", tkns.Message)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

func Reg(w http.ResponseWriter, r *http.Request) { //обработка Регистрации (POST)
	email := r.FormValue("email")
	em := govalidator.IsEmail(strings.TrimSpace(email)) //Проверка на форму email
	password := r.FormValue("password")
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //Хешируем
	tkns.CheckErr(err)

	name := r.FormValue("name")       //Имя
	surname := r.FormValue("surname") //Фамилия

	roleStr := r.FormValue("role")
	var author bool
	if roleStr == "true" { //Перевод из строки в bool
		author = true
	} else {
		author = false
	}

	if email == "" || password == "" || name == "" || surname == "" { //Проверка на заполнение формы
		tkns.Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	} else if em == false {
		tkns.Message = "Формат логина не соответсвует!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	var idd string
	erro := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&idd)
	if erro == nil { //Проверка на уникальность email
		tkns.Message = "Такой email уже зарегистрирован!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	}
	insert, err := db.Query("INSERT INTO users (email, passwordhash, name, surname, role) "+
		"VALUES ($1, $2, $3, $4, $5)", email, hashedPass, name, surname, author)
	tkns.CheckErr(err) //Сохраняем в бд
	insert.Close()
	http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
}
