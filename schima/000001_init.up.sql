CREATE TABLE users (
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    refresh_token TEXT,
    user_id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    refresh_token_expiry_time TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + '7 days'::interval)
);

CREATE TABLE articles (
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    post_id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author_id UUID NOT NULL,
    idempotency_key TEXT NOT NULL,
    status TEXT NOT NULL
);

CREATE TABLE pictures (
    image_id UUID NOT NULL PRIMARY KEY,
    post_id UUID NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    image_url TEXT NOT NULL
);