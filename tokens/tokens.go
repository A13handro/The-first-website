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
		return
	}
	Refreshtoken, ok := data["refresh_token"]
	if !ok {
		fmt.Println("refresh_token не найден в запросе: ", err)
		return
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	err = db.QueryRow("SELECT userid FROM users WHERE refreshtoken = $1", Refreshtoken).Scan(&UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	tokenR, err := ParseToken(Refreshtoken) //Проверяем валидность токена
	if tokenR.Valid {                       //Обновляем Accesstoken, если refreshtoken действителен
		Accesstoken, _ := GenerateToken(UserID, 2) // Обновляем токен на 2 часа

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    Accesstoken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   7200, // 2 часа
		})
		w.WriteHeader(http.StatusOK)
		return
	} else {
		fmt.Println("refreshtoken истек или недействителен")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc { //Основная проверка токена
	return func(w http.ResponseWriter, r *http.Request) {
		cookie1, err1 := r.Cookie("refresh_token")
		cookie2, err2 := r.Cookie("access_token")
		if err1 != nil || err2 != nil {
			fmt.Println("Токен не получен из печенек: ", err1, err2)
			w.WriteHeader(http.StatusBadRequest)
			jsonData, _ := json.Marshal(map[string]string{
				"Message": "Пользователь не авторизован",
			})
			w.Write(jsonData)
			return
		}
		Refreshtoken := cookie1.Value
		Accesstoken := cookie2.Value

		connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			fmt.Println("Ошибка: ", err)
			return
		}
		defer db.Close()

		tokenA, _ := ParseToken(Accesstoken) //Проверяем валидность токена
		if !tokenA.Valid {                   //Если токен не валиден, вызываем RefreshToken
			jsonData, err := json.Marshal(map[string]string{
				"refresh_token": Refreshtoken,
				"access_token":  Accesstoken,
			})
			if err != nil {
				fmt.Println("Ошибка JSON:", err)
				return
			}
			resp, err := http.Post("http://localhost:8080/api/auth/refresh-token", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Ошибка: ", err)
				return
			} else if resp.StatusCode != 200 {
				jsonData, _ := json.Marshal(map[string]string{
					"Message": "Пользователь не авторизован",
				})
				w.Write(jsonData)
				return
			}
			defer resp.Body.Close()
		}
		next(w, r)
	}
}
