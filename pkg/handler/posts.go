package handler

import (
	"encoding/json"
	"net/http"
	todo "the-first-website"

	"github.com/gorilla/mux"
)

// @Summary Создание поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Posts
// @Router /api/posts [post]
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var input todo.Post

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	Refreshtoken := cookie.Value

	err = h.services.Posts.CreatePost(Refreshtoken, input)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	newErrorResponse(w, http.StatusOK, "Пост успешно создан")
}

// @Summary Редактирование поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Posts
// @Router /api/posts/{postId} [put]
func (h *Handler) EditPost(w http.ResponseWriter, r *http.Request) {
	var input todo.Post
	PostId := mux.Vars(r)["postId"]

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	Refreshtoken := cookie.Value

	err = h.services.Posts.EditPost(Refreshtoken, input, PostId)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	newErrorResponse(w, http.StatusOK, "Пост успешно изменен")
}

// @Summary Публикация поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Posts
// @Router /api/posts/{postId}/status [PATCH]
func (h *Handler) PublishPost(w http.ResponseWriter, r *http.Request) {
	var input todo.Post
	PostId := mux.Vars(r)["postId"]

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	Refreshtoken := cookie.Value

	err = h.services.Posts.PublishPost(Refreshtoken, input, PostId)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	newErrorResponse(w, http.StatusOK, "Пост успешно опубликован")
}

// @Summary Просмотр постов
// @Description Требует аутентификации.
// @Tags Posts
// @Router /api/posts [Get]
func (h *Handler) ViewingPosts(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	Refreshtoken := cookie.Value

	jsonData, err := h.services.Posts.ViewingPosts(Refreshtoken)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
