package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
	"github.com/MAPiryazev/OzonTest/internal/service"
	"github.com/google/uuid"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// возвращает посты
func (h *Handler) ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	postList, err := h.svc.ListPosts(ctx, offset, limit)

	if err != nil {
		if errors.Is(err, customerrors.ErrParamOutOfRange) {
			return nil, err
		}
		return nil, fmt.Errorf("ошибка запроса к базе: %w", err)
	}
	return postList, nil
}

// возвращает пост по ID
func (h *Handler) GetPost(ctx context.Context, postID string) (*models.Post, error) {
	foundPost, err := h.svc.GetPostByID(ctx, postID)

	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("пост с id %s не найден", postID)
		}
		if errors.Is(err, customerrors.ErrValidation) {
			return nil, err
		}
		return nil, fmt.Errorf("не удалось получить пост: %w", err)
	}
	return foundPost, nil
}

// возвращает комментарии поста
func (h *Handler) ListComments(ctx context.Context, postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	commentList, err := h.svc.ListCommentsByPost(ctx, postID, parentID, offset, limit)

	if err != nil {
		if errors.Is(err, customerrors.ErrValidation) || errors.Is(err, customerrors.ErrParamOutOfRange) {
			return nil, err
		}
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("не удалось получить комментарии: %w", err)
	}
	return commentList, nil
}

// создает пользователя
func (h *Handler) CreateUser(ctx context.Context, username string) (*models.User, error) {
	newUser := &models.User{
		ID:       uuid.NewString(),
		Username: username}

	err := h.svc.CreateUser(ctx, newUser)

	if err != nil {
		if errors.Is(err, customerrors.ErrAlreadyExists) {
			return nil, customerrors.ErrAlreadyExists
		}
		if errors.Is(err, customerrors.ErrValidation) {
			return nil, err
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}
	return newUser, nil
}

// создает пост
func (h *Handler) CreatePost(ctx context.Context, title, content, creatorID string, commentsEnabled bool) (*models.Post, error) {
	newPost := &models.Post{
		ID:              uuid.NewString(),
		Title:           title,
		Content:         content,
		AuthorID:        creatorID,
		CommentsEnabled: commentsEnabled,
		CreatedAt:       time.Now().UTC(),
	}
	//распознавание ошибок
	if err := h.svc.CreatePost(ctx, newPost); err != nil {
		switch {
		case errors.Is(err, customerrors.ErrAlreadyExists):
			return nil, customerrors.ErrAlreadyExists
		case errors.Is(err, customerrors.ErrValidation):
			return nil, err
		case errors.Is(err, customerrors.ErrNotFound):
			return nil, fmt.Errorf("ошибка при создании поста: пользователь не найден: %w", customerrors.ErrEnvNotFound)
		default:
			return nil, fmt.Errorf("ошибка при создании поста: %w", err)
		}
	}

	return newPost, nil
}

// обновляет пост
func (h *Handler) UpdatePost(ctx context.Context, id, title, content, userID string) (*models.Post, error) {
	needUpdate := &models.Post{
		ID:      id,
		Title:   title,
		Content: content,
	}

	err := h.svc.UpdatePost(ctx, needUpdate, userID)

	if err != nil {
		if errors.Is(err, customerrors.ErrValidation) || errors.Is(err, customerrors.ErrForbidden) || errors.Is(err, customerrors.ErrNotFound) { //опознанные ошибки
			return nil, err
		}
		return nil, fmt.Errorf("ошибка обновления поста: %w", err)
	}

	return h.svc.GetPostByID(ctx, id)
}

// создает комментарий к посту
func (h *Handler) CreateComment(ctx context.Context, postID, text, authorID string, parentID *string) (*models.Comment, error) {
	comm := &models.Comment{
		ID:        uuid.NewString(),
		PostID:    postID,
		ParentID:  parentID,
		AuthorID:  authorID,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	err := h.svc.CreateComment(ctx, comm)
	if err != nil {
		if errors.Is(err, customerrors.ErrValidation) || errors.Is(err, customerrors.ErrCommForbidden) || errors.Is(err, customerrors.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("ошибка при создании комментария: %w", err)
	}

	return comm, nil
}
