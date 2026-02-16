package main

import (
	"os"
	todo "the-first-website"
	_ "the-first-website/docs"
	"the-first-website/pkg/handler"
	"the-first-website/pkg/repository"
	"the-first-website/pkg/service"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// @title API на Go со Swagger
// @version 1.0
// @description Сайт
// @contact.name Александр
// @contact.url https://vk.com/a13handro
// @host localhost:8080
// @BasePath /api/v1

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := InitConfig(); err != nil {
		logrus.Fatal("Ошибка configs: ", err)
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatal("Ошибка загрузки .env: ", err)
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     os.Getenv("HOST"),
		Port:     os.Getenv("PG_PORT"),
		Username: os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		DBName:   os.Getenv("PG_DATABASE"),
		SSLMode:  os.Getenv("PG_SSLMODE"),
	})
	if err != nil {
		logrus.Fatal("Ошибка db: ", err)
	}

	repos := repository.NewRepository(db)
	service := service.NewService(repos)
	handlers := handler.NewHandler(service)

	srv := new(todo.Server)
	if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		logrus.Fatal("Ошибка сервера: ", err)
	}
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
