package handler

import (
	"the-first-website/pkg/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *mux.Router {
	router := mux.NewRouter()

	api := router.PathPrefix("/api").Subrouter()
	auth := api.PathPrefix("/auth").Subrouter()
	posts := auth.PathPrefix("/posts").Subrouter()

	auth.HandleFunc("/register", h.Register).Methods("POST")
	auth.HandleFunc("/login", h.Login).Methods("POST")
	auth.HandleFunc("/refresh-token", h.RefreshToken).Methods("POST")

	auth.HandleFunc("/posts", h.AuthMiddleware(h.Viewing)).Methods("GET")
	auth.HandleFunc("/posts", h.AuthMiddleware(h.Posts)).Methods("POST")

	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/images",
		h.AuthMiddleware(h.Images)).Methods("POST")
	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}",
		h.AuthMiddleware(h.Edit)).Methods("PUT")
	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/images/"+
		"{imageId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}", h.AuthMiddleware(h.Delete)).Methods("DELETE")
	posts.HandleFunc("/{postId:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}}/status",
		h.AuthMiddleware(h.Publish)).Methods("PATCH")

	return router
}
