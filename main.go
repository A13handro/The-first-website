package main

import (
	"net/http"
	"the-first-website/auth"
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

	rtr.HandleFunc("/api/auth/register", auth.Register).Methods("POST")
	rtr.HandleFunc("/api/auth/login", auth.Login).Methods("POST")
	rtr.HandleFunc("/api/auth/refresh-token", tkns.RefreshToken).Methods("POST")
	rtr.HandleFunc("/api/posts", tkns.AuthMiddleware(pst.Posts)).Methods("POST")
	rtr.HandleFunc("/api/posts/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/images",
		tkns.AuthMiddleware(pst.Images)).Methods("POST")
	rtr.HandleFunc("/api/posts/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}",
		tkns.AuthMiddleware(pst.Edit)).Methods("PUT")
	rtr.HandleFunc("/api/posts/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/images/"+
		"{imageId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}", tkns.AuthMiddleware(pst.Delete)).Methods("DELETE")
	rtr.HandleFunc("/api/posts/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/status",
		tkns.AuthMiddleware(pst.Publish)).Methods("PATCH")
	rtr.HandleFunc("/api/posts", tkns.AuthMiddleware(mst.Viewing)).Methods("GET")

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)
	return rtr
}

func main() {
	handlfunc()
}
