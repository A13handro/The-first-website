package pst

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"os"

// 	"github.com/joho/godotenv"
// 	"github.com/minio/minio-go/v7"
// 	"github.com/minio/minio-go/v7/pkg/credentials"
// )

// func ServerMini(w http.ResponseWriter, r *http.Request) *minio.Client {
// 	err := godotenv.Load()
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		jsonData, _ := json.Marshal(map[string]string{
// 			"Message": "Ошибка загрузки .env",
// 			"Error":   err.Error(),
// 		})
// 		w.Write(jsonData)
// 	}
// 	//Запускаем minio
// 	ctx := context.Background()
// 	endpoint := os.Getenv("ENDPOINT")
// 	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
// 	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
// 	useSSL := false

// 	//Создаем клиент
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		jsonData, _ := json.Marshal(map[string]string{
// 			"Message": "Ошибка проверки бакета",
// 			"Error":   err.Error(),
// 		})
// 		w.Write(jsonData)
// 		return minioClient
// 	}
// 	// бакет
// 	bucketName := "pictures"
// 	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
// 	if err != nil {
// 		exists, err := minioClient.BucketExists(ctx, bucketName)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			jsonData, _ := json.Marshal(map[string]string{
// 				"Message": "Ошибка проверки бакета",
// 				"Error":   err.Error(),
// 			})
// 			w.Write(jsonData)
// 			return minioClient
// 		}
// 		if !exists {
// 			// Создаём бакет, если его нет
// 			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				jsonData, _ := json.Marshal(map[string]string{
// 					"Message": "Ошибка создания бакета",
// 					"Error":   err.Error(),
// 				})
// 				w.Write(jsonData)
// 				return minioClient
// 			}
// 			w.WriteHeader(http.StatusOK)
// 			jsonData, _ := json.Marshal(map[string]string{
// 				"Message": "Бакет успешно создан",
// 			})
// 			w.Write(jsonData)
// 		}
// 	}
// 	return minioClient
// }
