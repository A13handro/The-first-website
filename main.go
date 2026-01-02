package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Article struct {
	Id             uint16
	Title, Content string
	Data           string
	Mes            string
	Rol            bool
}

var Art = []Article{}
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
func parseToken(tokenString string) (*jwt.Token, error) { //Проверка токена на действительность
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}
func authMiddleware(next http.HandlerFunc) http.HandlerFunc { //Основная проверка токена
	return func(w http.ResponseWriter, r *http.Request) {
		connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		checkErr(err)
		defer db.Close()

		var authTokenA string //Переменная для коротковременного токена в бд
		var authTokenR string //Переменная для долговременного токена в бд

		table, err := db.Query("SELECT accesstoken FROM users WHERE id = $1", ThiseID)
		checkErr(err)
		if table.Next() {
			table.Scan(&authTokenA)
		}
		if AccessToken == "" || AccessToken != authTokenA { //Если нет сохраненного глобально токена
			Message = "Пожалуйста, авторизируйтесь!"
			http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
			return
		}
		tokenA, _ := parseToken(authTokenA) //Проверяем на действительность токен
		if !tokenA.Valid {
			err1 := db.QueryRow("SELECT refreshtoken FROM users WHERE id = $1", ThiseID).Scan(&authTokenR)
			checkErr(err1)                      //Вытягиваем refreshtoken
			tokenR, _ := parseToken(authTokenR) //Проверяем
			if tokenR.Valid {                   //Обновляем AccessToken, если refreshtoken действителен
				AccessToken, _ = generateToken(ThiseID, 15)
				_, err = db.Exec("UPDATE users SET accesstoken = $1 WHERE id = $2", AccessToken, ThiseID)
				checkErr(err)
			} else { //Иначе
				Message = "Время сеанса истекло. Пожалуйста, авторезируйтесь повторно."
				http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
				return
			}
		}
		next(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) { //Главная страница (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/index.html", "templates/footer.html")
	checkErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err1 := sql.Open("postgres", connStr)
	checkErr(err1)
	defer db.Close()

	table, err2 := db.Query("SELECT * FROM articles")
	checkErr(err2)

	var role bool
	var Art = []Article{}
	err3 := db.QueryRow("SELECT role FROM users WHERE id = $1", ThiseID).Scan(&role)
	checkErr(err3)
	for table.Next() { //Заполняем Art
		var post Article
		err = table.Scan(&post.Id, &post.Title, &post.Content, &post.Data)
		checkErr(err)
		post.Rol = role
		Art = append(Art, post)
	}
	data := struct { //Создаем структуру для передачи данных в html
		Posts []Article
		Role  bool
		Mes   string
	}{
		Posts: Art,
		Role:  role,
		Mes:   Message,
	}
	t.ExecuteTemplate(w, "index", data)
	Message = ""
}

func login(w http.ResponseWriter, r *http.Request) { //Авторизация (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/login.html")
	checkErr(err)

	AccessToken = "" // Выходим из аккаунта
	RefreshToken = ""

	t.ExecuteTemplate(w, "login", Message)
	Message = "" //Обнуляем сообщение после передачи пользователю
}

func log(w http.ResponseWriter, r *http.Request) { //обработка Авторизации (POST)
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" { //Проверка на заполнение формы
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkErr(err)
	defer db.Close()

	var storedHash string
	var id int
	err = db.QueryRow("SELECT passwordhash, id FROM users WHERE email = $1", email).Scan(&storedHash, &id) //Поиск
	if err != nil {
		Message = "Неверный логин"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) //Сравение пароля
	if err != nil {
		Message = "Неверный пароль"
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}

	ThiseID = id                             // Запоминаем ID пользователя
	AccessToken, _ = generateToken(id, 2)    // Создаем токен на 2 часа
	RefreshToken, _ = generateToken(id, 168) // Создаем токен на неделю
	_, err = db.Exec("UPDATE users SET accesstoken = $1, refreshtoken = $2 WHERE id = $3", AccessToken, RefreshToken, id)
	checkErr(err) //Отправляем токен в бд
	http.Redirect(w, r, "/api", http.StatusSeeOther)
	Message = "" //Обнуляем сообщение после передачи пользователю
}

func register(w http.ResponseWriter, r *http.Request) { //Регистрация (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/register.html")
	checkErr(err)

	AccessToken = "" // Выходим из аккаунта
	RefreshToken = ""

	t.ExecuteTemplate(w, "register", Message)
	Message = "" //Обнуляем сообщение после передачи пользователю
}

func reg(w http.ResponseWriter, r *http.Request) { //обработка Регистрации (POST)
	email := r.FormValue("email")
	em := govalidator.IsEmail(strings.TrimSpace(email)) //Проверка на форму email
	password := r.FormValue("password")
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //Хешируем
	checkErr(err)
	fmt.Println(hashedPass)

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
		Message = "Не все поля заполнены!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	} else if em == false {
		Message = "Формат логина не соответсвует!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkErr(err)
	defer db.Close()

	var idd string
	erro := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&idd)
	if erro == nil { //Проверка на уникальность email
		Message = "Такой email уже зарегистрирован!"
		http.Redirect(w, r, "/api/auth/register", http.StatusSeeOther)
		return
	}
	insert, err := db.Query("INSERT INTO users (email, passwordhash, name, surname, role) "+
		"VALUES ($1, $2, $3, $4, $5)", email, hashedPass, name, surname, author)
	checkErr(err) //Сохраняем в бд
	insert.Close()
	http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
}

func posts(w http.ResponseWriter, r *http.Request) { //Создание поста (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/posts.html", "templates/footer.html")
	checkErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkErr(err)
	defer db.Close()

	var role bool //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE id = $1", ThiseID).Scan(&role)
	checkErr(err)
	if role == false {
		Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}

	t.ExecuteTemplate(w, "posts", Message)
	Message = "" //Обнуляем сообщение после передачи пользователю
}

func pos(w http.ResponseWriter, r *http.Request) { //обработка создания или изменения поста (POST)
	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkErr(err)
	defer db.Close()

	flag := r.FormValue("pict")       //ФЛАЖОК, что картинки не будет
	title := r.FormValue("title")     //Название статьи
	content := r.FormValue("content") //Текст статьи
	idS := r.FormValue("id")          //ID
	id, _ := strconv.Atoi(idS)
	adres := r.FormValue("a")          //Адрес страницы, откуда вызвалась функция
	file, _, err := r.FormFile("file") //Файл картинки

	var pictur bool = false       //Здесь начинается сложная логика
	if flag == "" && err != nil { //НЕТ КАРТИНКИ и ФЛАЖКА
		pictur = true
	}

	if title == "" || content == "" || (pictur && adres == "/api/posts") { //Проверка на заполнение формы
		Message = "Не все поля заполнены!"
		if adres != "/api/posts" { //Отправляем обратно на ту же страницу
			http.Redirect(w, r, adres, http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/api/posts", http.StatusSeeOther)
		return
	}

	var fileCod string
	var fileBytes = []byte("s")        // Это значение будет обозначать отсутствие картинки
	if pictur == false && flag == "" { //Если есть картинка
		fileBytes, err = io.ReadAll(file)
		checkErr(err)
		defer file.Close()
	}
	fileCod = base64.StdEncoding.EncodeToString(fileBytes) //Кодируем для чтения в html

	ins, err3 := db.Query("SELECT * FROM articles WHERE id = $1", id)
	checkErr(err3)
	if ins.Next() { // Если пост редактируется
		if pictur == true { //Картинка не меняется
			_, err = db.Exec("UPDATE articles SET title = $1, content = $2 WHERE id = $3", title, content, id)
		} else { //Картинка обновляется
			_, err = db.Exec("UPDATE articles SET title = $1, content = $2, data = $3 WHERE id = $4", title, content, fileCod, id)
		}
		checkErr(err)
	} else { //Иначе просто создаем пост
		insert, err := db.Query("INSERT INTO articles (title, content, data) VALUES ($1, $2, $3)", title, content, fileCod)
		checkErr(err)
		insert.Close()
	}

	http.Redirect(w, r, "/api", http.StatusSeeOther)
	Message = "" //Обнуляем сообщение после передачи пользователю
}

func edit_post(w http.ResponseWriter, r *http.Request) { //Редактирование поста (GET)
	vars := mux.Vars(r) //Принимаем ID поста

	t, err := template.ParseFiles("templates/header.html", "templates/show.html", "templates/footer.html")
	checkErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkErr(err)
	defer db.Close()

	insert, err := db.Query("SELECT * FROM articles WHERE id = $1", vars["id"])
	checkErr(err)

	var role bool //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE id = $1", ThiseID).Scan(&role)
	checkErr(err)
	if role == false {
		Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}
	showPosts = Article{}
	for insert.Next() { //Читаем из бд каждый пост
		var post Article
		err = insert.Scan(&post.Id, &post.Title, &post.Content, &post.Data)
		post.Mes = Message
		checkErr(err)
		showPosts = post
	}

	t.ExecuteTemplate(w, "show", showPosts)
	Message = "" //Обнуляем сообщение после передачи пользователю
}

func delete_post(w http.ResponseWriter, r *http.Request) { //Удаление поста
	vars := mux.Vars(r)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkErr(err)
	defer db.Close()

	var role bool //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE id = $1", ThiseID).Scan(&role)
	checkErr(err)
	if role == false {
		Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}

	insert, err := db.Query("DELETE FROM articles WHERE id = $1", vars["id"])
	checkErr(err)
	insert.Close()

	http.Redirect(w, r, "/api", http.StatusSeeOther)
	Message = "" //Обнуляем сообщение после передачи пользователю
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
