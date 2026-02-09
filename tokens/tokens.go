package tkns

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func GenerateToken(userID uuid.UUID, num int) (string, error) { //Создание токена
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(time.Hour * time.Duration(num)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env: ", err)
	}
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	return token.SignedString(secretKey)
}

func ParseToken(tokenString string) (*jwt.Token, error) { //Проверка токена на действительность
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		return secretKey, nil
	})
}

// @Summary Обновление токена
// @Description метод вызывается автоматически из функции-оболочки (AuthMiddleware) при невалидном Accesstoken
// @Tags Authorization
// @Router /api/auth/refresh-token [post]
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println("Неверный JSON: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Неверный JSON",
			"Error":   err.Error(),
		})
		w.Write(jsonData)
		return
	}
	Refreshtoken, ok := data["refresh_token"]
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "refresh_token не найден в запросе",
		})
		w.Write(jsonData)
		return
	}

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Ошибка загрузки .env",
			"Error":   err.Error(),
		})
		w.Write(jsonData)
		return
	}
	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	err = db.QueryRow("SELECT user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	tokenR, err := ParseToken(Refreshtoken) //Проверяем валидность токена
	if tokenR.Valid {
		_, ok := data["Middleware"]
		if !ok { //Если функция вызвана не из AuthMiddleware
			Accesstoken, _ := GenerateToken(UserID, 2) // Обновляем токен на 2 часа
			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    Accesstoken,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   604800,
			})
		}
		w.WriteHeader(http.StatusOK)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "access_token обновлен",
		})
		w.Write(jsonData)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "refresh_token истек или недействителен",
		})
		w.Write(jsonData)
		return
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc { //Основная проверка токена
	return func(w http.ResponseWriter, r *http.Request) {
		cookie1, err1 := r.Cookie("refresh_token")
		cookie2, err2 := r.Cookie("access_token")
		if err1 != nil || err2 != nil {
			fmt.Println("Токен не получен из печенек: ", err1, err2)
			w.WriteHeader(http.StatusUnauthorized)
			jsonData, _ := json.Marshal(map[string]string{
				"Message": "Пользователь не авторизован",
			})
			w.Write(jsonData)
			return
		}
		Refreshtoken := cookie1.Value
		Accesstoken := cookie2.Value

		err := godotenv.Load()
		if err != nil {
			fmt.Println("Ошибка загрузки .env: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonData, _ := json.Marshal(map[string]string{
				"Message": "Ошибка загрузки .env",
				"Error":   err.Error(),
			})
			w.Write(jsonData)
			return
		}
		connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s",
			os.Getenv("PG_USER"),
			os.Getenv("PG_PASSWORD"),
			os.Getenv("PG_PORT"),
			os.Getenv("PG_DATABASE"),
			os.Getenv("PG_SSLMODE"),
		)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			fmt.Println("Ошибка: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonData, _ := json.Marshal(map[string]string{
				"Error": err.Error(),
			})
			w.Write(jsonData)
			return
		}
		defer db.Close()

		tokenA, _ := ParseToken(Accesstoken) //Проверяем валидность токена
		if !tokenA.Valid {                   //Если токен не валиден, вызываем RefreshToken
			jsonData, err := json.Marshal(map[string]string{
				"refresh_token": Refreshtoken,
				"Middleware":    "Middleware",
			})
			if err != nil {
				fmt.Println("Ошибка JSON:", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonData, _ := json.Marshal(map[string]string{
					"Message": "Ошибка JSON",
					"Error":   err.Error(),
				})
				w.Write(jsonData)
				return
			}
			resp, err := http.Post("http://localhost:8080/api/auth/refresh-token", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Ошибка: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonData, _ := json.Marshal(map[string]string{
					"Error": err.Error(),
				})
				w.Write(jsonData)
				return
			} else if resp.StatusCode != 200 {
				w.WriteHeader(http.StatusBadRequest)
				jsonData, _ := json.Marshal(map[string]string{
					"Message": "Пользователь не авторизован",
				})
				w.Write(jsonData)
				return
			}
			var UserID uuid.UUID
			err = db.QueryRow("SELECT user_id FROM users WHERE refresh_token = $1", Refreshtoken).Scan(&UserID)
			if err != nil {
				fmt.Println("Ошибка: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonData, _ := json.Marshal(map[string]string{
					"Error": err.Error(),
				})
				w.Write(jsonData)
				return
			}
			Accesstoken, _ := GenerateToken(UserID, 2)
			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    Accesstoken,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   604800,
			})
			w.WriteHeader(http.StatusOK)
			jsonData, _ = json.Marshal(map[string]string{
				"Message": "access_token обновлен",
			})
			w.Write(jsonData)
			defer resp.Body.Close()
		}
		next(w, r)
	}
}
