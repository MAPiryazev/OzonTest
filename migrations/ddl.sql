--модель данных для БД, которая автоматически создается при первом старте сервиса

create table users(
    id uuid primary key,
    username varchar(255) not null
);

create table posts(
    id uuid primary key,
    title varchar(255) not null,
    content text not null,
    author_id uuid not null references users(id),
    comments_enabled boolean default true,
    created_at timestamp not null default now()
);

CREATE TABLE comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id),
    parent_id UUID NULL REFERENCES comments(id),
    author_id UUID NOT NULL REFERENCES users(id),
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
