package pst

import (
	"database/sql"
	"encoding/base64"
	"html/template"
	"io"
	"net/http"
	tkns "the-first-website/tokens"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// @Summary Страница создания поста
// @Description Требует аутентификации. Проверяет роль. Если роль читателя, то выдает сообщение и перенаправляет на главную страницу.
// @Tags Posts
// @Router /api/posts [get]
func Posts(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("C:/Users/Александр/Desktop/The-first-website/templates/header.html", "C:/Users/Александр/Desktop/The-first-website/templates/posts.html", "C:/Users/Александр/Desktop/The-first-website/templates/footer.html")
	tkns.CheckErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	var role string //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE userid = $1", tkns.UserID).Scan(&role)
	tkns.CheckErr(err)
	if role == "Reader" {
		tkns.Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}

	t.ExecuteTemplate(w, "posts", tkns.Message)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

// @Summary Обработка создания поста или изменения поста
// @Description Требует аутентификации. Функция для создания и изменения поста. Сложная логика с картинками: при создании нужно выбрать будет картинка или нет, а при изменении можно не менять.
// @Tags Posts
// @Router /api/pos [post]
func Pos(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	flag := r.FormValue("pict")       //ФЛАЖОК, что картинки не будет
	title := r.FormValue("title")     //Название статьи
	content := r.FormValue("content") //Текст статьи
	idStr := r.FormValue("id")        //uuId
	id, _ := uuid.Parse(idStr)
	adres := r.FormValue("a")          //Адрес страницы, откуда вызвалась функция
	file, _, err := r.FormFile("file") //Файл картинки

	var pictur bool = false       //Здесь начинается сложная логика
	if flag == "" && err != nil { //НЕТ КАРТИНКИ и ФЛАЖКА
		pictur = true
	}

	if title == "" || content == "" || (pictur && adres == "/api/posts") { //Проверка на заполнение формы
		tkns.Message = "Не все поля заполнены!"
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
		tkns.CheckErr(err)
		defer file.Close()
	}
	fileCod = base64.StdEncoding.EncodeToString(fileBytes) //Кодируем для чтения в html

	ins, err3 := db.Query("SELECT * FROM articles WHERE postid = $1", id)
	tkns.CheckErr(err3)
	if ins.Next() { //Если пост редактируется
		if pictur == true { //Картинка не меняется
			_, err = db.Exec("UPDATE articles SET title = $1, content = $2, updatedat = $3 WHERE postid = $4", title, content, time.Now(), id)
		} else { //Картинка обновляется
			_, err = db.Exec("UPDATE articles SET title = $1, content = $2, updatedat = $3, images = $4 WHERE postid = $5", title, content, time.Now(), fileCod, id)
		}
		tkns.CheckErr(err)
	} else { //Иначе просто создаем пост
		insert, err := db.Query("INSERT INTO articles (title, content, images, createdat, updatedat, authorid) VALUES ($1, $2, $3, $4, $5, $6)", title, content, fileCod, time.Now(), time.Now(), tkns.UserID)
		tkns.CheckErr(err)
		insert.Close()
	}

	http.Redirect(w, r, "/api", http.StatusSeeOther)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

// @Summary Страница редактирования поста
// @Description Требует аутентификации.
// @Tags Posts
// @Router /api/post/{id} [get]
func Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) //Принимаем ID поста

	t, err := template.ParseFiles("templates/header.html", "templates/show.html", "templates/footer.html")
	tkns.CheckErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	insert, err := db.Query("SELECT * FROM articles WHERE postid = $1", vars["id"])
	tkns.CheckErr(err)
	var role string //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE userid = $1", tkns.UserID).Scan(&role)
	tkns.CheckErr(err)
	if role == "Reader" {
		tkns.Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}
	tkns.ShowPosts = tkns.Article{}
	for insert.Next() { //Читаем из бд каждый пост
		var post tkns.Article
		err = insert.Scan(&post.Title, &post.Content, &post.Images, &post.PostId, &post.Createdat, &post.Updatedat, &post.Authorid)
		post.Mes = tkns.Message
		tkns.CheckErr(err)
		tkns.ShowPosts = post
	}

	t.ExecuteTemplate(w, "show", tkns.ShowPosts)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

// @Summary Удаление поста
// @Description Требует аутентификации.
// @Tags Posts
// @Router /api/delete/{id} [get]
func Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	var role string //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE userid = $1", tkns.UserID).Scan(&role)
	tkns.CheckErr(err)
	if role == "Reader" {
		tkns.Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}
	insert, err := db.Query("DELETE FROM articles WHERE postid = $1", vars["id"])
	tkns.CheckErr(err)
	insert.Close()

	http.Redirect(w, r, "/api", http.StatusSeeOther)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}
