package main

import (
	"net/http"
	auth "the-first-website/auth"
	_ "the-first-website/docs"
	mst "the-first-website/master"
	pst "the-first-website/post"
	tkns "the-first-website/tokens"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title API на Go со Swagger
// @version 1.0
// @description Сайт
// @contact.name Александр
// @contact.url https://vk.com/a13handro
// @host localhost:8080
// @BasePath /api/v1

func handlfunc() *mux.Router {
	rtr := mux.NewRouter()

	rtr.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	rtr.HandleFunc("/api", tkns.AuthMiddleware(mst.Master)).Methods("GET")      //Главная страница
	rtr.HandleFunc("/api/posts", tkns.AuthMiddleware(pst.Posts)).Methods("GET") //Создание поста
	rtr.HandleFunc("/api/pos", pst.Pos).Methods("POST")                         //обработка Создания поста
	//Страница редактирования поста
	rtr.HandleFunc("/api/post/{id:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}", tkns.AuthMiddleware(pst.Edit)).Methods("GET")
	//Удаление поста
	rtr.HandleFunc("/api/delete/{id:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}", tkns.AuthMiddleware(pst.Delete)).Methods("GET")
	rtr.HandleFunc("/api/auth/login", auth.Login).Methods("GET")       //Авторизация
	rtr.HandleFunc("/api/auth/log", auth.Log).Methods("POST")          //обработка Авторизации
	rtr.HandleFunc("/api/auth/register", auth.Register).Methods("GET") //Регистрация
	rtr.HandleFunc("/api/auth/reg", auth.Reg).Methods("POST")          //обработка Регистрации

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)

	return rtr
}

func main() {
	handlfunc()
}
