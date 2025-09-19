package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
	"github.com/MAPiryazev/OzonTest/internal/repository"
)

const MAX_COMMENT_LENGTH = 2000

type Service interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)

	CreatePost(ctx context.Context, post *models.Post) error
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	ListPosts(ctx context.Context, offset int, limit int) ([]*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post, userID string) error

	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, id string) (*models.Comment, error)
	ListCommentsByPost(ctx context.Context, postID string, parentID *string, offset, limit int) ([]*models.Comment, error)
}

type service struct {
	repository repository.Storage
	cfg        *config.AppConfig
}

func NewService(repo repository.Storage, cfg *config.AppConfig) Service {
	return &service{repository: repo, cfg: cfg}
}

func (s *service) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("%w: пользователь не может быть nil", customerrors.ErrValidation)
	}

	user.ID = strings.TrimSpace(user.ID)
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	user.Username = strings.TrimSpace(user.Username)
	if len(user.Username) < s.cfg.MinUsernameLen {
		return fmt.Errorf("%w: Имя пользователя должно быть >= %d букв", customerrors.ErrValidation, s.cfg.MinUsernameLen)
	}

	return s.repository.CreateUser(ctx, user)
}

func (s *service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	trId := strings.TrimSpace(id)
	if trId == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}

	return s.repository.GetUserByID(ctx, trId)
}

// создает пост (и валидирует) и передает в БД
func (s *service) CreatePost(ctx context.Context, post *models.Post) error {
	if post == nil {
		return fmt.Errorf("%w: пост при создании не может быть nil", customerrors.ErrValidation)
	}
	post.ID = strings.TrimSpace(post.ID)
	if post.ID == "" {
		post.ID = uuid.NewString()
	}

	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)

	//вернуть ошибки
	if post.Title == "" {
		return fmt.Errorf("%w: обязательно нужен заголовок для создания поста", customerrors.ErrValidation)
	}
	if post.Content == "" {
		return fmt.Errorf("%w: обязательно нужен контент для создания поста", customerrors.ErrValidation)
	}
	if post.AuthorID == "" {
		return fmt.Errorf("%w: нужен id автора", customerrors.ErrValidation)
	}

	_, err := s.GetUserByID(ctx, post.AuthorID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return customerrors.ErrNotFound
		}
		return err
	}
	if post.CreatedAt.IsZero() {
		post.CreatedAt = time.Now().UTC()
	}

	return s.repository.CreatePost(ctx, post)
}

func (s *service) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	trId := strings.TrimSpace(id)
	if trId == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}

	currPost, err := s.repository.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("ошибка при получении поста: %w", err)
	}

	return currPost, nil
}

func (s *service) ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	if offset < 0 || limit <= 0 || limit > s.cfg.MaxListLimit {
		return nil, fmt.Errorf("%w: неправильные параметры пагинации", customerrors.ErrParamOutOfRange)
	}

	return s.repository.ListPosts(ctx, offset, limit)
}

func (s *service) UpdatePost(ctx context.Context, post *models.Post, userID string) error {
	if post == nil {
		return fmt.Errorf("%w: пост при обновлении не может быть nil", customerrors.ErrValidation)
	}

	trPostID := strings.TrimSpace(post.ID)
	trUserID := strings.TrimSpace(userID)

	if trPostID == "" || trUserID == "" {
		return fmt.Errorf("%w: id поста и пользователя обязательны", customerrors.ErrValidation)
	}

	trPostTitle := strings.TrimSpace(post.Title)
	trPostContent := strings.TrimSpace(post.Content)

	if trPostTitle == "" || trPostContent == "" {
		return fmt.Errorf("%w: title и content обязательны", customerrors.ErrValidation)
	}

	currentPost, err := s.repository.GetPostByID(ctx, post.ID)
	if err != nil {
		return fmt.Errorf("%w: пост не найден %s: %v", customerrors.ErrNotFound, post.ID, err)
	}

	if currentPost.AuthorID != userID {
		return fmt.Errorf("%w: запрещено редактировать чужой пост", customerrors.ErrForbidden)
	}

	return s.repository.UpdatePost(ctx, post)
}

// создает комментарий и валидирует его
func (s *service) CreateComment(ctx context.Context, comment *models.Comment) error {
	if comment == nil {
		return fmt.Errorf("%w: комментарий при создании не может быть nil", customerrors.ErrValidation)
	}
	if comment.ID == "" {
		comment.ID = uuid.NewString()
	}

	trCommentText := strings.TrimSpace(comment.Text)
	if trCommentText == "" {
		return fmt.Errorf("%w: текст комментария не может быть пустым", customerrors.ErrValidation)
	}
	if len(comment.Text) > MAX_COMMENT_LENGTH {
		return fmt.Errorf("%w: длина комментария больше %d символов", customerrors.ErrParamOutOfRange, MAX_COMMENT_LENGTH)
	}

	trCommentPostID := strings.TrimSpace(comment.PostID)
	if trCommentPostID == "" {
		return fmt.Errorf("%w: Id поста для комментария не может быть пустым", customerrors.ErrValidation)
	}

	if comment.ParentID != nil {
		trParentID := strings.TrimSpace(*comment.ParentID)
		if trParentID == "" {
			return fmt.Errorf("%w: parentID пустой", customerrors.ErrValidation)
		}
		parentComment, err := s.repository.GetCommentByID(ctx, trParentID)
		if err != nil {
			return fmt.Errorf("%w: parent comment %s: %v", customerrors.ErrNotFound, trParentID, err)
		}
		if parentComment.PostID != comment.PostID {
			return fmt.Errorf("%w: parent comment %s принадлежит другому посту", customerrors.ErrValidation, trParentID)
		}
		comment.ParentID = &trParentID
	}

	post, err := s.repository.GetPostByID(ctx, comment.PostID)
	if err != nil {
		return fmt.Errorf("%w: пост не найден %s: %v", customerrors.ErrNotFound, comment.PostID, err)
	}
	if !post.CommentsEnabled {
		return fmt.Errorf("%w: комментарии к посту запрещены ", customerrors.ErrCommForbidden)
	}

	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now().UTC()
	}

	return s.repository.CreateComment(ctx, comment)
}

func (s *service) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	trId := strings.TrimSpace(id)

	if trId == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}

	return s.repository.GetCommentByID(ctx, id)
}

func (s *service) ListCommentsByPost(ctx context.Context, postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return nil, fmt.Errorf("%w: Id для получения не может быть пустым", customerrors.ErrValidation)
	}

	if offset < 0 || limit <= 0 || limit > s.cfg.MaxListLimit {
		return nil, fmt.Errorf("%w: неправильные параметры пагинации", customerrors.ErrParamOutOfRange)
	}

	// проверка на существование поста
	if _, err := s.repository.GetPostByID(ctx, postID); err != nil {
		return nil, fmt.Errorf("%w: пост %s: %v", customerrors.ErrNotFound, postID, err)
	}

	// проверка parentID
	if parentID != nil {
		trParentID := strings.TrimSpace(*parentID)
		if trParentID == "" {
			return nil, fmt.Errorf("%w: parentID пустой", customerrors.ErrValidation)
		}
		parentComment, err := s.repository.GetCommentByID(ctx, trParentID)
		if err != nil {
			return nil, fmt.Errorf("%w: parent comment %s: %v", customerrors.ErrNotFound, trParentID, err)
		}
		if parentComment.PostID != postID {
			return nil, fmt.Errorf("%w: parent comment %s не соответствует посту", customerrors.ErrValidation, trParentID)
		}
		p := trParentID
		parentID = &p
	}

	return s.repository.ListCommentsByPost(ctx, postID, parentID, offset, limit)
}
