package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
	"github.com/MAPiryazev/OzonTest/internal/service"
	"github.com/google/uuid"
)

type Handler struct {
	svc service.Service
	cfg *config.AppConfig
}

func NewHandler(svc service.Service, cfg *config.AppConfig) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

func (h *Handler) ListPosts(offset, limit int) ([]*models.Post, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > h.cfg.MaxListLimit {
		limit = h.cfg.DefaultListLimit
	}
	posts, err := h.svc.ListPosts(offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return posts, nil
}

func (h *Handler) GetPost(id string) (*models.Post, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id пустое", customerrors.ErrValidation)
	}
	post, err := h.svc.GetPostByID(id)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return post, nil
}

func (h *Handler) ListComments(postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	if postID == "" {
		return nil, fmt.Errorf("%w: id поста для получения пустое", customerrors.ErrValidation)
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > h.cfg.MaxListLimit {
		limit = h.cfg.DefaultListLimit
	}
	comments, err := h.svc.ListCommentsByPost(postID, parentID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return comments, nil
}

func (h *Handler) CreateUser(username string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if len(username) < h.cfg.MinUsernameLen {
		return nil, fmt.Errorf("%w: Имя пользователя должно быть >= %d букв", customerrors.ErrValidation, h.cfg.MinUsernameLen)
	}
	user := &models.User{
		ID:       uuid.NewString(),
		Username: username,
	}
	err := h.svc.CreateUser(user)
	if err != nil {
		if errors.Is(err, customerrors.ErrAlreadyExists) {
			return nil, customerrors.ErrAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return user, nil
}

func (h *Handler) CreatePost(title, content, authorId string, commentsEnabled bool) (*models.Post, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("%w: обязательно нужен загловок для создания поста", customerrors.ErrValidation)
	}
	if authorId == "" {
		return nil, fmt.Errorf("%w: нужен id автора", customerrors.ErrValidation)
	}
	_, err := h.svc.GetUserByID(authorId)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: пользователь не найден", customerrors.ErrEnvNotFound)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	post := &models.Post{
		ID:              uuid.NewString(),
		Title:           title,
		Content:         content,
		AuthorID:        authorId,
		CommentsEnabled: commentsEnabled,
		CreatedAt:       time.Now().UTC(),
	}
	err = h.svc.CreatePost(post)
	if err != nil {
		if errors.Is(err, customerrors.ErrAlreadyExists) {
			return nil, customerrors.ErrAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return post, nil
}

func (h *Handler) UpdatePost(id, title, content, userId string) (*models.Post, error) {
	if id == "" || userId == "" {
		return nil, fmt.Errorf("%w: id поста и пользователя обязательны", customerrors.ErrValidation)
	}
	existing, err := h.svc.GetPostByID(id)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	if existing.AuthorID != userId {
		return nil, customerrors.ErrForbidden
	}
	updated := &models.Post{
		ID:      id,
		Title:   title,
		Content: content,
	}
	if err := h.svc.UpdatePost(updated, userId); err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return h.svc.GetPostByID(id)
}

func (h *Handler) CreateComment(postId, text, authorId string, parentId *string) (*models.Comment, error) {
	if postId == "" || authorId == "" {
		return nil, fmt.Errorf("%w: id поста и автора обязательны", customerrors.ErrValidation)
	}
	if len(text) == 0 {
		return nil, fmt.Errorf("%w: у комментария должен быть текст", customerrors.ErrValidation)
	}
	if len(text) > h.cfg.MaxCommentLength {
		return nil, fmt.Errorf("%w: длина комментария дожна быть <= %d", customerrors.ErrParamOutOfRange, h.cfg.MaxCommentLength)
	}

	post, err := h.svc.GetPostByID(postId)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	if !post.CommentsEnabled {
		return nil, customerrors.ErrCommForbidden
	}

	if parentId != nil {
		parent, err := h.svc.GetCommentByID(*parentId)
		if err != nil {
			if errors.Is(err, customerrors.ErrNotFound) {
				return nil, customerrors.ErrNotFound
			}
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
		}
		if parent.PostID != postId {
			return nil, fmt.Errorf("%w: родитель принадлежит другому посту", customerrors.ErrValidation)
		}
	}

	comment := &models.Comment{
		ID:        uuid.NewString(),
		PostID:    postId,
		ParentID:  parentId,
		AuthorID:  authorId,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}
	err = h.svc.CreateComment(comment)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return comment, nil
}
