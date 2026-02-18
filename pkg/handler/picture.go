package handler

import (
	"net/http"
	todo "the-first-website"

	"github.com/gorilla/mux"
)

// @Summary Добавление картинки к посту
// @Description Требует аутентификации(Роль Автор).
// @Tags Image
// @Router /api/posts/{postId}/images [post]
func (h *Handler) AddImage(w http.ResponseWriter, r *http.Request) {
	PostId := mux.Vars(r)["postId"]
	var input todo.Image
	file, Header, err := r.FormFile("image")
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	input.File = file
	input.FileHeader = *Header

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	Refreshtoken := cookie.Value

	err = h.services.Picture.AddImage(Refreshtoken, input, PostId)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	newErrorResponse(w, http.StatusOK, "Картинка успешно добавлена")
}

// @Summary Удаление картинки из поста
// @Description Требует аутентификации(Роль Автор).
// @Tags Image
// @Router /api/posts/{postId}/images/{imageId} [delete]
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) //postId
	PostId := vars["postId"]
	ImageId := vars["imageId"]

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		newErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	Refreshtoken := cookie.Value

	err = h.services.Picture.DeleteImage(Refreshtoken, ImageId, PostId)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	newErrorResponse(w, http.StatusOK, "Картинка успешно удалена")

}
