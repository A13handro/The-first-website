package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	tkns "the-first-website/tokens"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
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
		return
	}

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	defer db.Close()

	var storedHash string
	var UserID uuid.UUID
	err = db.QueryRow("SELECT passwordhash, userid FROM users WHERE email = $1", email).Scan(&storedHash, &UserID)

	err1 := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) //Сравение пароля

	if err != nil || err1 != nil {
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
	_, err = db.Exec("UPDATE users SET refreshtoken = $1, refreshtokenexpirytime = CURRENT_TIMESTAMP + INTERVAL '7 days' WHERE userid = $2", Refreshtoken, UserID)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    Refreshtoken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   604800, // неделя
	})
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

	connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	defer db.Close()

	var UserID uuid.UUID
	erro := db.QueryRow("SELECT userid FROM users WHERE email = $1", email).Scan(&UserID)
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

	insert, err := db.Query("INSERT INTO users (email, passwordhash, role, refreshtoken) "+
		"VALUES ($1, $2, $3, $4)", email, hashedPass, role, Refreshtoken)
	if err != nil {
		fmt.Println("Ошибка: ", err)
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
		MaxAge:   604800, // неделя
	})
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
	jsonData, _ := json.Marshal(map[string]string{
		"Message": "Регистрация успешна",
	})
	w.Write(jsonData)
}
