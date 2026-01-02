package tkns

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type Article struct {
	Id             uint16
	Title, Content string
	Data           string
	Mes            string
	Rol            bool
}

var Art = []Article{}
var ShowPosts = Article{}

var AccessToken string
var RefreshToken string

var ThiseID int
var Message string
var secretKey = []byte("my_secret_key")

func CheckErr(err error) { //Проверка на ошибку
	if err != nil {
		Message = "Ошибка"
		fmt.Println("Ошибка:")
		panic(err)
	}
}

func GenerateToken(userID int, num int) (string, error) { //Создание токена
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(time.Hour * time.Duration(num)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ParseToken(tokenString string) (*jwt.Token, error) { //Проверка токена на действительность
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc { //Основная проверка токена
	return func(w http.ResponseWriter, r *http.Request) {
		connStr := "user=postgres password=123 port=5432 dbname=usersdb sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		CheckErr(err)
		defer db.Close()

		var authTokenA string //Переменная для коротковременного токена в бд
		var authTokenR string //Переменная для долговременного токена в бд

		table, err := db.Query("SELECT accesstoken FROM users WHERE id = $1", ThiseID)
		CheckErr(err)
		if table.Next() {
			table.Scan(&authTokenA)
		}
		if AccessToken == "" || AccessToken != authTokenA { //Если нет сохраненного глобально токена
			Message = "Пожалуйста, авторизируйтесь!"
			http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
			return
		}
		tokenA, _ := ParseToken(authTokenA) //Проверяем на действительность токен
		if !tokenA.Valid {
			err1 := db.QueryRow("SELECT refreshtoken FROM users WHERE id = $1", ThiseID).Scan(&authTokenR)
			CheckErr(err1)                      //Вытягиваем refreshtoken
			tokenR, _ := ParseToken(authTokenR) //Проверяем
			if tokenR.Valid {                   //Обновляем AccessToken, если refreshtoken действителен
				AccessToken, _ = GenerateToken(ThiseID, 15)
				_, err = db.Exec("UPDATE users SET accesstoken = $1 WHERE id = $2", AccessToken, ThiseID)
				CheckErr(err)
			} else { //Иначе
				Message = "Время сеанса истекло. Пожалуйста, авторезируйтесь повторно."
				http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
				return
			}
		}
		next(w, r)
	}
}
