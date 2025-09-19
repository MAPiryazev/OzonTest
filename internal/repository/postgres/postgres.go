package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
	_ "github.com/lib/pq"
)

// реализация интерфейса storage как хранилища в postgres

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(cfg *config.DBConfig) (*PostgresStorage, error) {
	dbCredentials := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", dbCredentials)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBCreation, err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBMaxConnLifeTime) * time.Minute)

	// проверяем соединение с базой
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ошибка при проверке соединения с БД: %v", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

// проверяет есть ли пользователь в БД и если нет, то добавляет
func (p *PostgresStorage) CreateUser(ctx context.Context, user *models.User) error {
	trimmedName := strings.TrimSpace(user.Username)

	var exists bool
	checkQuery := `select exists(select 1 from users where username=$1)`
	if err := p.db.QueryRow(checkQuery, trimmedName).Scan(&exists); err != nil {
		return fmt.Errorf("ошибка проверки существования пользователя: %w", err)
	}
	if exists {
		return fmt.Errorf("пользователь %s уже существует: %w", trimmedName, customerrors.ErrAlreadyExists)
	}

	insertQuery := `insert into users (id, username) values ($1,$2)`
	if _, err := p.db.Exec(insertQuery, user.ID, trimmedName); err != nil {
		return fmt.Errorf("не удалось создать пользователя с id %s: %w", user.ID, err)
	}

	return nil
}

func (p *PostgresStorage) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `select id, username from users where id = $1`
	row := p.db.QueryRow(query, id)

	var u models.User
	err := row.Scan(&u.ID, &u.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: пользователь с id %s", customerrors.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return &u, nil
}

func (p *PostgresStorage) CreatePost(ctx context.Context, post *models.Post) error {
	if post.CreatedAt.IsZero() {
		post.CreatedAt = time.Now()
	}
	query := `insert into posts (id, title, content, author_id, comments_enabled, created_at) values ($1,$2,$3,$4,$5,$6)`
	_, err := p.db.Exec(query, post.ID, post.Title, post.Content, post.AuthorID, post.CommentsEnabled, post.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка при создании поста с id %s: %w", post.ID, err)
	}
	return nil
}

func (p *PostgresStorage) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	query := `select id, title, content, author_id, comments_enabled, created_at from posts where id = $1`
	row := p.db.QueryRow(query, id)

	var currPost models.Post
	err := row.Scan(&currPost.ID, &currPost.Title, &currPost.Content, &currPost.AuthorID, &currPost.CommentsEnabled, &currPost.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: пост с id %s", customerrors.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return &currPost, nil
}

func (p *PostgresStorage) ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильные параметры пагинации", customerrors.ErrParamOutOfRange)
	}

	query := `select id, title, content, author_id, comments_enabled, created_at from posts order by created_at desc offset $1 limit $2`
	rows, err := p.db.Query(query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CommentsEnabled, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

// обновляет пост
func (p *PostgresStorage) UpdatePost(ctx context.Context, post *models.Post) error {
	query := `update posts set title=$1, content=$2, comments_enabled=$3 where id=$4`
	res, err := p.db.Exec(query, post.Title, post.Content, post.CommentsEnabled, post.ID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении поста: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: пост с id %s", customerrors.ErrNotFound, post.ID)
	}
	return nil
}

// добавляет комментарий в БД, провалидирован в service
func (p *PostgresStorage) CreateComment(ctx context.Context, comment *models.Comment) error {
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now()
	}
	query := `insert into comments (id, post_id, parent_id, author_id, text, created_at) values ($1,$2,$3,$4,$5,$6)`
	_, err := p.db.Exec(query, comment.ID, comment.PostID, comment.ParentID, comment.AuthorID, comment.Text, comment.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка при создании комментария с id %s : %w", comment.ID, err)
	}
	return nil
}

func (p *PostgresStorage) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	query := `select id, post_id, parent_id, author_id, text, created_at from comments where id = $1`
	row := p.db.QueryRow(query, id)

	var currComment models.Comment
	err := row.Scan(&currComment.ID, &currComment.PostID, &currComment.ParentID, &currComment.AuthorID, &currComment.Text, &currComment.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: комментарий с id %s", customerrors.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return &currComment, nil
}

func (p *PostgresStorage) ListCommentsByPost(ctx context.Context, postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильные параметры пагинации", customerrors.ErrParamOutOfRange)
	}

	var rows *sql.Rows
	var err error

	if parentID == nil {
		query := `select id, post_id, parent_id, author_id, text, created_at from comments
				where post_id = $1 and parent_id is null
				order by created_at asc
				offset $2 limit $3`
		rows, err = p.db.Query(query, postID, offset, limit)
	} else {
		query := `select id, post_id, parent_id, author_id, text, created_at from comments
				where post_id = $1 and parent_id = $2
				order by created_at asc
				offset $3 limit $4`
		rows, err = p.db.Query(query, postID, *parentID, offset, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.AuthorID, &comment.Text, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}
