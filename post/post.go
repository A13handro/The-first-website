package pst

import (
	"database/sql"
	"encoding/base64"
	"html/template"
	"io"
	"net/http"
	"strconv"
	tkns "the-first-website/tokens"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func Posts(w http.ResponseWriter, r *http.Request) { //Создание поста (GET)
	t, err := template.ParseFiles("templates/header.html", "templates/posts.html", "templates/footer.html")
	tkns.CheckErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	var role bool //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE id = $1", tkns.ThiseID).Scan(&role)
	tkns.CheckErr(err)
	if role == false {
		tkns.Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}

	t.ExecuteTemplate(w, "posts", tkns.Message)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

func Pos(w http.ResponseWriter, r *http.Request) { //обработка создания или изменения поста (POST)
	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
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

	ins, err3 := db.Query("SELECT * FROM articles WHERE id = $1", id)
	tkns.CheckErr(err3)
	if ins.Next() { // Если пост редактируется
		if pictur == true { //Картинка не меняется
			_, err = db.Exec("UPDATE articles SET title = $1, content = $2 WHERE id = $3", title, content, id)
		} else { //Картинка обновляется
			_, err = db.Exec("UPDATE articles SET title = $1, content = $2, data = $3 WHERE id = $4", title, content, fileCod, id)
		}
		tkns.CheckErr(err)
	} else { //Иначе просто создаем пост
		insert, err := db.Query("INSERT INTO articles (title, content, data) VALUES ($1, $2, $3)", title, content, fileCod)
		tkns.CheckErr(err)
		insert.Close()
	}

	http.Redirect(w, r, "/api", http.StatusSeeOther)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

func Edit_post(w http.ResponseWriter, r *http.Request) { //Редактирование поста (GET)
	vars := mux.Vars(r) //Принимаем ID поста

	t, err := template.ParseFiles("templates/header.html", "templates/show.html", "templates/footer.html")
	tkns.CheckErr(err)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	insert, err := db.Query("SELECT * FROM articles WHERE id = $1", vars["id"])
	tkns.CheckErr(err)

	var role bool //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE id = $1", tkns.ThiseID).Scan(&role)
	tkns.CheckErr(err)
	if role == false {
		tkns.Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}
	tkns.ShowPosts = tkns.Article{}
	for insert.Next() { //Читаем из бд каждый пост
		var post tkns.Article
		err = insert.Scan(&post.Id, &post.Title, &post.Content, &post.Data)
		post.Mes = tkns.Message
		tkns.CheckErr(err)
		tkns.ShowPosts = post
	}

	t.ExecuteTemplate(w, "show", tkns.ShowPosts)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}

func Delete_post(w http.ResponseWriter, r *http.Request) { //Удаление поста
	vars := mux.Vars(r)

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	tkns.CheckErr(err)
	defer db.Close()

	var role bool //Проверка на права
	err = db.QueryRow("SELECT role FROM users WHERE id = $1", tkns.ThiseID).Scan(&role)
	tkns.CheckErr(err)
	if role == false {
		tkns.Message = "У вас нет прав!"
		http.Redirect(w, r, "/api", http.StatusSeeOther)
		return
	}

	insert, err := db.Query("DELETE FROM articles WHERE id = $1", vars["id"])
	tkns.CheckErr(err)
	insert.Close()

	http.Redirect(w, r, "/api", http.StatusSeeOther)
	tkns.Message = "" //Обнуляем сообщение после передачи пользователю
}
