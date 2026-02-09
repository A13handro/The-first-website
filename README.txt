Инструкции по запуску тестов:

1. Создайте файл .env по образцу .env-example

3. Создаём таблицу для пользователей:

CREATE TABLE users (
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    refresh_token TEXT,
    user_id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    refresh_token_expiry_time TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + '7 days'::interval)
)

4. Создаём таблицу для постов:

CREATE TABLE articles (
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    post_id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author_id UUID NOT NULL,
    idempotency_key TEXT NOT NULL,
    status TEXT NOT NULL
)

5. Создаём таблицу для картинок:

CREATE TABLE pictures (
    image_id UUID NOT NULL PRIMARY KEY,
    post_id UUID NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    image_url TEXT NOT NULL
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