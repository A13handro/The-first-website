package handler

import (
	"the-first-website/pkg/service"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	api := router.PathPrefix("/api").Subrouter()
	auth := api.PathPrefix("/auth").Subrouter()
	posts := api.PathPrefix("/posts").Subrouter()

	auth.HandleFunc("/register", h.Register).Methods("POST")
	auth.HandleFunc("/login", h.Login).Methods("POST")
	auth.HandleFunc("/refresh-token", h.RefreshToken).Methods("POST")

	api.HandleFunc("/posts", h.AuthMiddleware(h.ViewingPosts)).Methods("GET")
	api.HandleFunc("/posts", h.AuthMiddleware(h.CreatePost)).Methods("POST")

	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/images",
		h.AuthMiddleware(h.AddImage)).Methods("POST")
	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}",
		h.AuthMiddleware(h.EditPost)).Methods("PUT")
	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/images/"+
		"{imageId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}", h.AuthMiddleware(h.DeleteImage)).Methods("DELETE")
	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/status",
		h.AuthMiddleware(h.PublishPost)).Methods("PATCH")

	return router
}
