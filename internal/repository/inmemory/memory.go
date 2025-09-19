package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
)

// in-memory хранилище
type MemoryStorage struct {
	mu          sync.RWMutex
	usersByID   map[string]*models.User
	usersByName map[string]*models.User
	posts       map[string]*models.Post
	comments    map[string]*models.Comment
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		usersByID:   make(map[string]*models.User),
		usersByName: make(map[string]*models.User),
		posts:       make(map[string]*models.Post),
		comments:    make(map[string]*models.Comment),
	}
}

// создаёт пользователя с  проверкой уникальности имени
func (m *MemoryStorage) CreateUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.usersByName[user.Username]; exists {
		return fmt.Errorf("%w: пользователь с именем %s уже существует", customerrors.ErrAlreadyExists, user.Username)
	}

	m.usersByID[user.ID] = user
	m.usersByName[user.Username] = user
	return nil
}

func (m *MemoryStorage) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.usersByID[id]
	if !ok {
		return nil, fmt.Errorf("%w: пользователь с id %s", customerrors.ErrNotFound, id)
	}
	return user, nil
}

func (m *MemoryStorage) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	post, ok := m.posts[id]
	if !ok {
		return nil, fmt.Errorf("%w: пост с id %s", customerrors.ErrNotFound, id)
	}
	return post, nil
}

// возвращает список постов
func (m *MemoryStorage) ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильный параметр пагинации", customerrors.ErrParamOutOfRange)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	allPosts := make([]*models.Post, 0, len(m.posts))
	for _, val := range m.posts {
		allPosts = append(allPosts, val)
	}

	if offset >= len(allPosts) {
		return []*models.Post{}, nil
	}

	end := offset + limit
	if end > len(allPosts) {
		end = len(allPosts)
	}
	return allPosts[offset:end], nil
}

func (m *MemoryStorage) CreatePost(ctx context.Context, post *models.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.posts[post.ID]; exists {
		return fmt.Errorf("%w: пост с id %s уже существует", customerrors.ErrAlreadyExists, post.ID)
	}

	post.CreatedAt = time.Now().UTC()
	m.posts[post.ID] = post
	return nil
}

// обновляет пост
func (m *MemoryStorage) UpdatePost(ctx context.Context, post *models.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.posts[post.ID]; !exists {
		return fmt.Errorf("%w: пост с id %s не найден", customerrors.ErrNotFound, post.ID)
	}

	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) CreateComment(ctx context.Context, comment *models.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.comments[comment.ID]; exists {
		return fmt.Errorf("%w: комментарий с id %s уже существует", customerrors.ErrAlreadyExists, comment.ID)
	}

	comment.CreatedAt = time.Now().UTC()
	m.comments[comment.ID] = comment
	return nil
}

func (m *MemoryStorage) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	comment, ok := m.comments[id]
	if !ok {
		return nil, fmt.Errorf("%w: комментарий с id %s не найден", customerrors.ErrNotFound, id)
	}
	return comment, nil
}

func (m *MemoryStorage) ListCommentsByPost(ctx context.Context, postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильный параметр пагинации", customerrors.ErrParamOutOfRange)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	result := []*models.Comment{}
	for _, val := range m.comments {
		if val.PostID != postID {
			continue
		}
		if parentID == nil && val.ParentID == nil {
			result = append(result, val)
		} else if parentID != nil && val.ParentID != nil && *parentID == *val.ParentID {
			result = append(result, val)
		}
	}

	if offset >= len(result) {
		return []*models.Comment{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

// Close здесь просто чтобы интерфейс был реализован
func (m *MemoryStorage) Close() error {
	return nil
}
