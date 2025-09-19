package repository

import (
	"context"

	"github.com/MAPiryazev/OzonTest/internal/models"
)

// Интерфейс хранилища для взаимодействия с объектами (psql or in-memory)
type Storage interface {
	CreatePost(ctx context.Context, post *models.Post) error
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post) error

	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, id string) (*models.Comment, error)
	ListCommentsByPost(ctx context.Context, postID string, parentID *string, offset, limit int) ([]*models.Comment, error)

	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)

	Close() error
}
