Инструкции по запуску тестов:

1. Создайте файл .env по образцу .env-example

2. Создаем базу данных:

CREATE database usersdb

3. Создаём таблицу для пользователей:

CREATE TABLE users (
    email TEXT NOT NULL,
    passwordhash TEXT NOT NULL,
    role TEXT NOT NULL,
    refreshtoken TEXT,
    userid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    refreshtokenexpirytime TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + '7 days'::interval)
)

4. Создаём таблицу для постов:

CREATE TABLE articles (
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    postid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    createdat TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedat TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    authorid UUID NOT NULL,
    idempotencykey TEXT NOT NULL,
    status TEXT NOT NULL
)

5. Создаём таблицу для картинок:

CREATE TABLE pictures (
    imageid UUID NOT NULL PRIMARY KEY,
    postid UUID NOT NULL,
    createdat TIMESTAMP WITHOUT TIME ZONE NOT NULL
)

6. Запускаем в терминале:

    go run main.go

6. Документация расположена по адресу:

    http://localhost:8080/swagger/index.html

7. Чтобы зайти на сайт нужно перейти по адресу:

Чтобы протестировать функции через postman импортируйте
 Tests.postman_collection.json
из проекта в postman и измените адреса: uuid постов и картинок
и файл картинки, когда будете ее добавлять