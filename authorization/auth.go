package authorization

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	tkns "the-first-website/tokens"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Авторизация
// @Tags Authorization
// @Router /api/auth/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		fmt.Println("Не все поля заполнены")
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Не все поля заполнены",
		})
		w.Write(jsonData)
		return
	}

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

	var storedHash string
	var UserID uuid.UUID
	err = db.QueryRow("SELECT password_hash, user_id FROM users WHERE email = $1", email).Scan(&storedHash, &UserID)
	if err != nil {
		fmt.Println("Неверный логин или пароль")
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Неверный логин или пароль",
		})
		w.Write(jsonData)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) //Сравение пароля
	if err != nil {
		fmt.Println("Неверный логин или пароль")
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Неверный логин или пароль",
		})
		w.Write(jsonData)
		return
	}

	Accesstoken, _ := tkns.GenerateToken(UserID, 2)    // Создаем токен на 2 часа
	Refreshtoken, _ := tkns.GenerateToken(UserID, 168) // Создаем токен на неделю
	_, err = db.Exec("UPDATE users SET refresh_token = $1, refresh_token_expiry_time = CURRENT_TIMESTAMP + INTERVAL '7 days' WHERE user_id = $2", Refreshtoken, UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    Refreshtoken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})
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
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Вход успешен",
	})
	w.Write(jsonData)

}

// @Summary Регистрация
// @Tags Authorization
// @Router /api/auth/register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	IsEmail := govalidator.IsEmail(strings.TrimSpace(email))
	if IsEmail == false {
		fmt.Println("Неверный формат email")
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Неверный формат email",
		})
		w.Write(jsonData)
		return
	}
	password := r.FormValue("password")
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //Хешируем
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	role := r.FormValue("role") //Author или Reader

	if email == "" || password == "" || role == "" { //Проверка на заполнение формы
		fmt.Println("Не все поля заполнены")
		w.WriteHeader(http.StatusBadRequest)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Не все поля заполнены",
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
	erro := db.QueryRow("SELECT user_id FROM users WHERE email = $1", email).Scan(&UserID)
	if erro == nil {
		fmt.Println("Email уже существует")
		w.WriteHeader(http.StatusForbidden)
		jsonData, _ := json.Marshal(map[string]string{
			"Message": "Email уже существует",
		})
		w.Write(jsonData)
		return
	}
	UserID = uuid.New()
	Accesstoken, _ := tkns.GenerateToken(UserID, 2)    // Создаем токен на 2 часа
	Refreshtoken, _ := tkns.GenerateToken(UserID, 168) // Создаем токен на неделю
	refreshtokenexpirytime := time.Now().Add(7 * 24 * time.Hour)
	insert, err := db.Query("INSERT INTO users (email, password_hash, role, refresh_token, refresh_token_expiry_time) "+
		"VALUES ($1, $2, $3, $4, $5)", email, hashedPass, role, Refreshtoken, refreshtokenexpirytime)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonData, _ := json.Marshal(map[string]string{
			"Error": err.Error(),
		})
		w.Write(jsonData)
		return
	}
	insert.Close()

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    Refreshtoken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800,
	})
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
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Регистрация успешна",
	})
	w.Write(jsonData)
}
