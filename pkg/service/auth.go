package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"strings"
	todo "the-first-website"
	"the-first-website/pkg/repository"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const salt = "arsdfhj213ipo98f"

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(user todo.User) (uuid.UUID, error) {
	user.Password = generatePasswordHash(user.Password)
	user.IsEmail = govalidator.IsEmail(strings.TrimSpace(user.Email))
	return s.repo.CreateUser(user)
}

func (s *AuthService) CheckUser(user todo.User) (uuid.UUID, error) {
	UserID, pass, err := s.repo.CheckUser(user)
	if err != nil {
		return UserID, err
	}
	user.Password = generatePasswordHash(user.Password)
	if user.Password != pass {
		return UserID, errors.New("пароль не совпадает")
	}
	return UserID, nil
}

func (s *AuthService) GenerateToken(UserID uuid.UUID, num int) (string, error) {
	claims := jwt.MapClaims{
		"id":  UserID,
		"exp": time.Now().Add(time.Hour * time.Duration(num)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if err := godotenv.Load(); err != nil {
		return "", errors.New("Ошибка загрузки .env")
	}
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	return token.SignedString(secretKey)
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password + salt))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (*AuthService) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if err := godotenv.Load(); err != nil {
			return "", errors.New("Ошибка загрузки .env")
		}
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		return secretKey, nil
	})
}

func (s *AuthService) GetID(refresh_token string) (uuid.UUID, error) {
	return s.repo.GetID(refresh_token)
}
