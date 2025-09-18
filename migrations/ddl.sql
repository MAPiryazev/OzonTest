--простоая модель данных, которая автоматически создается при первом старте сервиса и покрывает нужды тестового задания
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

create table comments (
    id uuid primary key,
    post_id uuid not null references posts(id),
    parent_id uuid null references comments(id), --связь с родительским комментарием
    author_id uuid not null references users(id),
    text varchar(2000) not null, --2к символов ограничение
    created_at timestamp not null default now()
);

--индексы
create unique index idx_users_username on users(username);
create index idx_posts_author_id on posts(author_id);
create index idx_posts_created_at on posts(created_at);
create index idx_comments_post_id on comments(post_id);
create index idx_comments_parent_id on comments(parent_id);
create index idx_comments_author_id on comments(author_id);
create index idx_comments_created_at on comments(created_at);
