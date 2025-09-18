package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
)

// реализация интерфейса storage как in-memory хранилище
type MemoryStorage struct {
	mu       sync.RWMutex
	users    map[string]*models.User
	posts    map[string]*models.Post
	comments map[string]*models.Comment
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:    make(map[string]*models.User),
		posts:    make(map[string]*models.Post),
		comments: make(map[string]*models.Comment),
	}
}

func (m *MemoryStorage) CreateUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.users[user.Username]
	if ok {
		return fmt.Errorf("%w: пользователь с именем %s", customerrors.ErrAlreadyExists, user.Username)
	}

	m.users[user.Username] = user
	return nil
}

func (m *MemoryStorage) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[id]
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

func (m *MemoryStorage) ListPosts(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильный параметр пагинации", customerrors.ErrParamOutOfRange)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	allPosts := make([]*models.Post, 0, limit)
	for _, val := range m.posts {
		allPosts = append(allPosts, val)
	}

	if offset > len(allPosts) {
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

	_, ok := m.posts[post.ID]
	if ok {
		return fmt.Errorf("%w: пост с id %s", customerrors.ErrAlreadyExists, post.ID)
	}

	post.CreatedAt = time.Now()
	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) UpdatePost(ctx context.Context, post *models.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.posts[post.ID]
	if !ok {
		return fmt.Errorf("%w: пост с id %s", customerrors.ErrNotFound, post.ID)
	}
	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) CreateComment(ctx context.Context, comment *models.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.comments[comment.ID]
	if ok {
		return fmt.Errorf("%w: комментарий с  id %s", customerrors.ErrAlreadyExists, comment.ID)

	}
	comment.CreatedAt = time.Now()
	m.comments[comment.ID] = comment
	return nil
}

func (m *MemoryStorage) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	comment, ok := m.comments[id]
	if !ok {
		return nil, fmt.Errorf("%w: комментарий с id %s", customerrors.ErrNotFound, id)
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
		// фильтруем по ID поста
		if val.PostID == postID {
			// если parentID не указан — берем только корневые комментарии
			if parentID == nil && val.ParentID == nil {
				result = append(result, val)
				// если parentID указан — берем только комментарии с этим parentID
			} else if parentID != nil && val.ParentID != nil && *parentID == *val.ParentID {
				result = append(result, val)
			}
		}
	}

	if offset > len(result) {
		return []*models.Comment{}, nil
	}

	stop := offset + limit
	if stop > len(result) {
		stop = len(result)
	}
	return result[offset:stop], nil
}
