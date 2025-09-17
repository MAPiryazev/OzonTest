package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
	"github.com/MAPiryazev/OzonTest/internal/repository"
)

const MAX_COMMENT_LENGTH = 2000

type Service interface {
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)

	CreatePost(post *models.Post) error
	GetPostByID(id string) (*models.Post, error)
	ListPosts(offset int, limit int) ([]*models.Post, error)
	UpdatePost(post *models.Post, userID string) error

	CreateComment(comment *models.Comment) error
	GetCommentByID(id string) (*models.Comment, error)
	ListCommentsByPost(postID string, parentID *string, offset, limit int) ([]*models.Comment, error)
}

type service struct {
	repository repository.Storage
}

func NewService(repo repository.Storage) Service {
	return &service{repository: repo}
}

func (s *service) CreateUser(user *models.User) error {
	if user == nil {
		return fmt.Errorf("%w: пользователь не может быть nil", customerrors.ErrValidation)
	}
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	user.Username = strings.TrimSpace(user.Username)
	if user.Username == "" {
		return fmt.Errorf("%w: имя пользователя не может быть пустым", customerrors.ErrValidation)
	}
	return s.repository.CreateUser(user)
}

func (s *service) GetUserByID(id string) (*models.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}
	return s.repository.GetUserByID(id)
}

func (s *service) CreatePost(post *models.Post) error {
	if post == nil {
		return fmt.Errorf("%w: пост при создании не может быть nil", customerrors.ErrValidation)
	}
	if post.ID == "" {
		post.ID = uuid.NewString()
	}
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)
	if post.Title == "" || post.Content == "" {
		return fmt.Errorf("%w: title и content обязательны при создании поста", customerrors.ErrValidation)
	}
	if post.CreatedAt.IsZero() {
		post.CreatedAt = time.Now()
	}
	return s.repository.CreatePost(post)
}

func (s *service) GetPostByID(id string) (*models.Post, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}
	return s.repository.GetPostByID(id)
}

func (s *service) ListPosts(offset, limit int) ([]*models.Post, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильные параметры пагинации", customerrors.ErrParamOutOfRange)
	}
	return s.repository.ListPosts(offset, limit)
}

func (s *service) UpdatePost(post *models.Post, userID string) error {
	if post == nil {
		return fmt.Errorf("%w: пост при обновлении не может быть nil", customerrors.ErrValidation)
	}
	post.ID = strings.TrimSpace(post.ID)
	if post.ID == "" {
		return fmt.Errorf("%w: Id при обновленни поста не может быть пустым", customerrors.ErrValidation)
	}
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)
	if post.Title == "" || post.Content == "" {
		return fmt.Errorf("%w: title и content обязательны", customerrors.ErrValidation)
	}

	existingPost, err := s.repository.GetPostByID(post.ID)
	if err != nil {
		return fmt.Errorf("%w: пост не найден %s: %v", customerrors.ErrNotFound, post.ID, err)
	}

	if existingPost.AuthorID != userID {
		return fmt.Errorf("%w: нельзя редактировать чужой пост", customerrors.ErrForbidden)
	}

	return s.repository.UpdatePost(post)
}

func (s *service) CreateComment(comment *models.Comment) error {
	if comment == nil {
		return fmt.Errorf("%w: комментарий при создании не может быть nil", customerrors.ErrValidation)
	}
	if comment.ID == "" {
		comment.ID = uuid.NewString()
	}
	comment.Text = strings.TrimSpace(comment.Text)
	if comment.Text == "" {
		return fmt.Errorf("%w: текст комментария не может быть пустым", customerrors.ErrValidation)
	}
	if len(comment.Text) > MAX_COMMENT_LENGTH {
		return fmt.Errorf("%w: длина комментария больше %d символов", customerrors.ErrValidation, MAX_COMMENT_LENGTH)
	}

	if comment.ID == "" {
		return fmt.Errorf("%w: Id поста для комментария не может быть пустым", customerrors.ErrValidation)
	}

	if comment.ParentID != nil {
		parentID := strings.TrimSpace(*comment.ParentID)
		if parentID == "" {
			return fmt.Errorf("%w: parentID пустой", customerrors.ErrValidation)
		}
		parentComment, err := s.repository.GetCommentByID(parentID)
		if err != nil {
			return fmt.Errorf("%w: parent comment %s: %v", customerrors.ErrNotFound, parentID, err)
		}
		if parentComment.PostID != comment.PostID {
			return fmt.Errorf("%w: parent comment %s принадлежит другому посту", customerrors.ErrValidation, parentID)
		}
	}

	post, err := s.repository.GetPostByID(comment.PostID)
	if err != nil {
		return fmt.Errorf("%w: пост не найден %s: %v", customerrors.ErrNotFound, comment.PostID, err)
	}
	if !post.CommentsEnabled {
		return fmt.Errorf("%w: комментарии запрещены автором", customerrors.ErrCommForbidden)
	}

	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now()
	}

	return s.repository.CreateComment(comment)
}

func (s *service) GetCommentByID(id string) (*models.Comment, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}
	return s.repository.GetCommentByID(id)
}

func (s *service) ListCommentsByPost(postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильные параметры пагинации", customerrors.ErrParamOutOfRange)
	}
	// проверка на существование поста
	if _, err := s.repository.GetPostByID(postID); err != nil {
		return nil, fmt.Errorf("%w: пост %s: %v", customerrors.ErrNotFound, postID, err)
	}

	// если задан parentID — убедиться, что он существует и принадлежит данному посту
	if parentID != nil {
		pidTrim := strings.TrimSpace(*parentID)
		if pidTrim == "" {
			return nil, fmt.Errorf("%w: parentID пустой", customerrors.ErrValidation)
		}
		parentComment, err := s.repository.GetCommentByID(pidTrim)
		if err != nil {
			return nil, fmt.Errorf("%w: parent comment %s: %v", customerrors.ErrNotFound, pidTrim, err)
		}
		if parentComment.PostID != postID {
			return nil, fmt.Errorf("%w: parent comment %s принадлежит другому посту", customerrors.ErrValidation, pidTrim)
		}
		p := pidTrim
		parentID = &p
	}

	return s.repository.ListCommentsByPost(postID, parentID, offset, limit)
}
