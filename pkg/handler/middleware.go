package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie1, err1 := r.Cookie("refresh_token")
		cookie2, err2 := r.Cookie("access_token")
		if err1 != nil || err2 != nil {
			newErrorResponse(w, http.StatusUnauthorized, err1.Error()+" "+err2.Error())
			return
		}
		Refreshtoken := cookie1.Value
		Accesstoken := cookie2.Value

		TokenA, err := h.services.Authorization.ParseToken(Accesstoken)
		if err != nil {
			newErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !TokenA.Valid {
			jsonData, err := json.Marshal(map[string]string{
				"refresh_token": Refreshtoken,
				"Middleware":    "Middleware",
			})
			if err != nil {
				newErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}

			if err := godotenv.Load(); err != nil {
				newErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			host := os.Getenv("JWT_SECRET")
			port := viper.GetString("port")

			resp, err := http.Post(fmt.Sprintf("http://%s:%s/api/auth/refresh-token", host, port), "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				newErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			} else if resp.StatusCode != 200 {
				newErrorResponse(w, http.StatusBadRequest, "Пользователь не авторизован")
				return
			}
			defer resp.Body.Close()

			UserID, err := h.services.Authorization.GetID(Refreshtoken)
			if err != nil {
				newErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			AccessToken, err := h.services.Authorization.GenerateToken(UserID, 2)
			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    AccessToken,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   604800,
			})

			newErrorResponse(w, http.StatusOK, "Пользователь авторизован")
		}
		next(w, r)
	}
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	Refreshtoken, ok := data["refresh_token"]
	if !ok {
		newErrorResponse(w, http.StatusInternalServerError, "refresh_token не найден в запросе")
		return
	}

	TokenR, err := h.services.Authorization.ParseToken(Refreshtoken)
	if TokenR.Valid {
		_, ok := data["Middleware"]
		if !ok {
			UserID, err := h.services.Authorization.GetID(Refreshtoken)
			if err != nil {
				newErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			AccessToken, err := h.services.Authorization.GenerateToken(UserID, 2)
			if err != nil {
				newErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    AccessToken,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   604800,
			})
		}
		newErrorResponse(w, http.StatusOK, "access_token обновлен")
		return
	} else {
		newErrorResponse(w, http.StatusInternalServerError, "refresh_token истек или недействителен")
		return
	}

}
