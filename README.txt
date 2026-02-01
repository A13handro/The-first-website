Инструкции по запуску сайта:

1. Создаем базу данных:

    CREATE database usersdb

2. Создаём таблицу пользователей:

    CREATE TABLE users (
        email CHARACTER VARYING(200) NOT NULL,
        passwordhash CHARACTER VARYING(200) NOT NULL,
        name CHARACTER VARYING(200) NOT NULL,
        surname CHARACTER VARYING(200) NOT NULL,
        accesstoken CHARACTER VARYING(512),
        role CHARACTER VARYING(6) NOT NULL,
        refreshtoken CHARACTER VARYING(512),
        userid UUID NOT NULL DEFAULT gen_random_uuid(),
        PRIMARY KEY (userid)
    )

3. Создаём таблицу для постов:

    CREATE TABLE articles (
        title CHARACTER VARYING(50) NOT NULL,
        content TEXT NOT NULL,
        images BYTEA DEFAULT '\x73'::BYTEA,
        postid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
        createdat TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updatedat TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        authorid UUID
    )

4. Запускаем в терминале:

    go run main.go

5. Документация расположена по адресу:

    http://localhost:8080/swagger/index.html

6. Чтобы зайти на сайт нужно перейти по адресу:

    http://localhost:8080/api/auth/register

7. Сначала нужно зарегестрироваться, потом войти. На главной странице постов пока нет,
их можно добавить(если при регистрации выбрана роль автора), перейдя на страницу создания постов.
Создать пост можно с картинкой или без. При редактировании поста картинку можно не менять,
можно убрать или выбрать новую. Посты могут изменять все авторы. Читатели могут только смотреть посты.
Также можно выйти из аккаунта.