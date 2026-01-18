package main

import (
	"net/http"
	auth "the-first-website/auth"
	mst "the-first-website/master"
	pst "the-first-website/post"
	tkns "the-first-website/tokens"

	_ "the-first-website/docs"

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

func handlfunc() {
	rtr := mux.NewRouter()

	rtr.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	rtr.HandleFunc("/api", tkns.AuthMiddleware(mst.Master)).Methods("GET")      //Главная страница
	rtr.HandleFunc("/api/posts", tkns.AuthMiddleware(pst.Posts)).Methods("GET") //Создание поста
	rtr.HandleFunc("/api/pos", pst.Pos).Methods("POST")                         //обработка Создания поста
	rtr.HandleFunc("/post/{id:[0-9]+}", tkns.AuthMiddleware(pst.Edit_post)).Methods("GET")
	rtr.HandleFunc("/delete_post/{id:[0-9]+}", tkns.AuthMiddleware(pst.Delete_post)).Methods("GET") //Удаление поста
	rtr.HandleFunc("/api/auth/login", auth.Login).Methods("GET")                                    //Авторизация
	rtr.HandleFunc("/api/auth/log", auth.Log).Methods("POST")                                       //обработка Авторизации
	rtr.HandleFunc("/api/auth/register", auth.Register).Methods("GET")                              //Регистрация
	rtr.HandleFunc("/api/auth/reg", auth.Reg).Methods("POST")                                       //обработка Регистрации

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handlfunc()
}
