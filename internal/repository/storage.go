package repository

import (
	"github.com/MAPiryazev/OzonTest/internal/models"
)

// Интерфейс хранилища для взаимодействия с объектами (psql or in-memory)
type Storage interface {
	CreatePost(post *models.Post) error
	GetPostByID(id string) (*models.Post, error)
	ListPosts(offset, limit int) ([]*models.Post, error)
	UpdatePost(post *models.Post) error

	CreateComment(comment *models.Comment) error
	GetCommentByID(id string) (*models.Comment, error)
	ListCommentsByPost(postID string, parentID *string, offset, limit int) ([]*models.Comment, error)

	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
}
