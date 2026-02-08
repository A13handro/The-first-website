package pst

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ServerMini(w http.ResponseWriter, r *http.Request) *minio.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Ошибка загрузки .env: ", err)
	}
	//Запускаем minio
	ctx := context.Background()
	endpoint := "localhost:9000"
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	//Создаем клиент
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}
	// бакет
	bucketName := "pictures"
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err != nil {
			log.Fatalln("Ошибка проверки бакета:", err)
		}
		if !exists {
			// Создаём бакет, если его нет
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Fatalln("Ошибка создания бакета:", err)
			}
			log.Printf("Бакет %s создан", bucketName)
		}
	}

	// // Получаем файл
	// file, header, err := r.FormFile("image")
	// if err != nil {
	// 	fmt.Println("Ошибка получения файла: ", err)
	// }
	// defer file.Close()

	// objectName := header.Filename
	// contentType := header.Header.Get("Content-Type")
	// if contentType == "" {
	// 	contentType = "application/octet-stream"
	// }

	// //Выгружаем файл в minio
	// _, err = minioClient.PutObject(
	// 	context.Background(),
	// 	bucketName,
	// 	objectName,
	// 	file,
	// 	header.Size,
	// 	minio.PutObjectOptions{ContentType: contentType},
	// )
	// if err != nil {
	// 	fmt.Println("Ошибка загрузки в MinIO: ", err)
	// }

	// log.Printf("Файл %s загружен в бакет %s\n", objectName, bucketName)
	// w.WriteHeader(http.StatusOK)
	// jsonData, _ := json.Marshal(map[string]string{
	// 	"Message": "Файл успешно загружен в бакет",
	// })
	// w.Write(jsonData)
	return minioClient
}
