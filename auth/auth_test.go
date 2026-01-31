package auth

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	tkns "the-first-website/tokens"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

func TestLogin(t *testing.T) {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/auth/login", Login).Methods("GET")

	// http запрос
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)

	// Создаём recorder для захвата ответа сервера
	w := httptest.NewRecorder()

	// Вызываем  функцию
	Register(w, req)

	// Результат из recorder
	resp := w.Result()
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("ожидается Content-Type text/plain, получено %s", contentType)
	}
	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ошибка чтения тела ответа: %v", err)
	}
	bodyStr := string(body)

	// Проверяем шаблон
	if !strings.Contains(bodyStr, "<title>Регистрация</title>") {
		t.Error("В ответе отсутствует ожидаемый html")
	}
}

func TestLog(t *testing.T) {
	// Тестовая БД
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Таблица users
	_, err = db.Exec(`
		CREATE TABLE users (
			email TEXT UNIQUE NOT NULL,
			passwordhash TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		nameForm    string
		email       string
		password    string
		expectedMsg string
		setupDB     func()
	}{
		{
			nameForm:    "Успешный вход",
			email:       "user@example.com",
			password:    "securepass123",
			expectedMsg: "",
			setupDB: func() {
				db.Exec("INSERT INTO users (email, password) VALUES (?, ?)",
					"existing@example.com", "securepass123")
			}},
		{
			nameForm:    "Пустые поля",
			email:       "",
			password:    "",
			expectedMsg: "Не все поля заполнены!",
			setupDB:     nil,
		},
		{
			nameForm:    "Неверный пароль",
			email:       "existing@example.com",
			password:    "securepass123",
			expectedMsg: "Неверный пароль!",
			setupDB: func() {
				db.Exec("INSERT INTO users (email, password) VALUES (?, ?)",
					"existing@example.com", "securepass1234")
			},
		},
		{
			nameForm:    "Неверный email",
			email:       "invalid-email",
			password:    "securepass123",
			expectedMsg: "Неверный email!",
			setupDB: func() {
				db.Exec("INSERT INTO users (email, password) VALUES (?, ?)",
					"existing@example.com", "securepass123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.nameForm, func(t *testing.T) {
			// Очищаем Message перед каждым тестом
			tkns.Message = ""

			if tt.setupDB != nil {
				tt.setupDB()
			}

			// http запрос
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(
				"email="+tt.email+
					"&password="+tt.password,
			))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()

			// Вызываем функцию
			Log(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			// Проверяем куда редиректим
			location := resp.Header.Get("Location")
			switch {
			case tt.expectedMsg == "Не все поля заполнены!":
				if location != "/api/auth/login" {
					t.Errorf("got redirect to %s, want /api/auth/login", location)
				}
			case tt.expectedMsg == "Неверный пароль!":
				if location != "/api/auth/login" {
					t.Errorf("got redirect to %s, want /api/auth/login", location)
				}
			case tt.expectedMsg == "Неверный email!":
				if location != "/api/auth/login" {
					t.Errorf("got redirect to %s, want /api/auth/login", location)
				}
			default:
				if location != "/api" {
					t.Errorf("got redirect to %s, want /api/auth/login", location)
				}
			}
		})
	}
}

func TestRegister(t *testing.T) {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/auth/register", Register).Methods("GET")

	// http запрос
	req := httptest.NewRequest(http.MethodGet, "/api/auth/register", nil)

	// Создаём recorder для захвата ответа сервера
	w := httptest.NewRecorder()

	// Вызываем  функцию
	Register(w, req)

	// Результат из recorder
	resp := w.Result()
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("ожидается Content-Type text/plain, получено %s", contentType)
	}
	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ошибка чтения тела ответа: %v", err)
	}
	bodyStr := string(body)

	// Проверяем шаблон
	if !strings.Contains(bodyStr, "<title>Регистрация</title>") {
		t.Error("В ответе отсутствует ожидаемый html")
	}
}

func TestReg(t *testing.T) {
	// Тестовая БД
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Таблица users
	_, err = db.Exec(`
		CREATE TABLE users (
			userid TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			passwordhash TEXT NOT NULL,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			role TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		nameForm    string
		email       string
		password    string
		name        string
		surname     string
		role        string
		expectedMsg string
		setupDB     func()
	}{
		{
			nameForm:    "Успешная регистрация",
			email:       "user@example.com",
			password:    "securepass123",
			name:        "Alice",
			surname:     "Smith",
			role:        "on",
			expectedMsg: "",
			setupDB:     nil,
		},
		{
			nameForm:    "Пустые поля",
			email:       "",
			password:    "",
			name:        "",
			surname:     "",
			role:        "",
			expectedMsg: "Не все поля заполнены!",
			setupDB:     nil,
		},
		{
			nameForm:    "Неверный email",
			email:       "invalid-email",
			password:    "pass",
			name:        "Bob",
			surname:     "Jones",
			role:        "off",
			expectedMsg: "Формат логина не соответсвует!",
			setupDB:     nil,
		},
		{
			nameForm:    "Дубликат email",
			email:       "existing@example.com",
			password:    "pass",
			name:        "Charlie",
			surname:     "Brown",
			role:        "on",
			expectedMsg: "Такой email уже зарегистрирован!",
			setupDB: func() {
				db.Exec("INSERT INTO users (email) VALUES (?)", "existing@example.com")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.nameForm, func(t *testing.T) {
			// Очищаем Message перед каждым тестом
			tkns.Message = ""

			if tt.setupDB != nil {
				tt.setupDB()
			}

			// http запрос
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(
				"email="+tt.email+
					"&password="+tt.password+
					"&name="+tt.name+
					"&surname="+tt.surname+
					"&role="+tt.role,
			))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()

			// Вызываем функцию
			Reg(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			// Проверяем куда редиректим
			location := resp.Header.Get("Location")
			switch {
			case tt.expectedMsg == "Не все поля заполнены!":
				if location != "/api/auth/register" {
					t.Errorf("got redirect to %s, want /api/auth/register", location)
				}
			case tt.expectedMsg == "Формат логина не соответсвует!":
				if location != "/api/auth/register" {
					t.Errorf("got redirect to %s, want /api/auth/register", location)
				}
			case tt.expectedMsg == "Такой email уже зарегистрирован!":
				if location != "/api/auth/register" {
					t.Errorf("got redirect to %s, want /api/auth/register", location)
				}
			}
			// Проверяем, что пользователь добавлен в БД
			if tt.expectedMsg == "" && tt.nameForm == "Успешная регистрация!" {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", tt.email).Scan(&count)
				if err != nil {
					t.Error("failed to query DB:", err)
				}
				if count != 1 {
					t.Error("user not found in DB")
				}
			}
		})
	}
}
