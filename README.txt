Инструкции по запуску:

1. Создайте файл .env по образцу .env-example

2. Создаём таблицы с помощью миграций:

migrate -path ./schima -database 'postgres://[PG_USER]:[PG_PASSWORD]@[HOST]:[PG_PORT]/[PG_DATABASE]?sslmode=[PG_SSLMODE]' up

Только вместо [PG_...] указать свои данные из вашего файла .env, созданного в первом пункте инструкции

3. Если понадобится, можно также удалить таблицы с помощью аналогичной миграции:

migrate -path ./schima -database 'postgres://[PG_USER]:[PG_PASSWORD]@[HOST]:[PG_PORT]/[PG_DATABASE]?sslmode=[PG_SSLMODE]' down

4. Запускаем в терминале:

    go run cmd/main.go 

5. Документация расположена по адресу:

    http://localhost:8080/swagger/index.html

6. Чтобы протестировать функции через postman импортируйте
Postman Collection.postman_collection.json
из проекта в postman и измените адреса: uuid постов, картинок
и файл картинки, когда будете ее добавлять