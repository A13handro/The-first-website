package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	todo "the-first-website"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input todo.User

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(input)
		newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	UserID, err := h.services.Authorization.CreateUser(input)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error()) //Всегда один код, пока не нашел решения этого
		return
	}

	AccessToken, err := h.services.Authorization.GenerateToken(UserID, 2)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	RefreshToken, err := h.services.Authorization.GenerateToken(UserID, 168)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    RefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    AccessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})

	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"Message":      "Регистрация успешна",
		"Id":           UserID.String(),
		"RefreshToken": RefreshToken,
		"AccessToken":  AccessToken,
	})
	w.Write(jsonData)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input todo.User

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	UserID, err := h.services.Authorization.CheckUser(input)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error()) //Всегда один код(500), пока не нашел решения этого
		return
	}

	AccessToken, err := h.services.Authorization.GenerateToken(UserID, 2)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	RefreshToken, err := h.services.Authorization.GenerateToken(UserID, 168)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    RefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    AccessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})
	newErrorResponse(w, http.StatusOK, "Авторизация успешна")
}
