package inmemory

import (
	"fmt"
	"sync"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
)

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

func (m *MemoryStorage) CreateUser(user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.users[user.ID]
	if exists {
		return fmt.Errorf("%w: пользователь с id %s", customerrors.ErrAlreadyExists, user.ID)
	}
	m.users[user.ID] = user
	return nil
}

func (m *MemoryStorage) GetUserByID(id string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("%w: пользователь с id %s", customerrors.ErrNotFound, id)
	}
	return user, nil
}

func (m *MemoryStorage) GetPostByID(id string) (*models.Post, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	post, ok := m.posts[id]
	if !ok {
		return nil, fmt.Errorf("%w: пост с id %s", customerrors.ErrNotFound, id)
	}
	return post, nil
}

func (m *MemoryStorage) ListPosts(offset, limit int) ([]*models.Post, error) {
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

func (m *MemoryStorage) CreatePost(post *models.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.posts[post.ID]
	if exists {
		return fmt.Errorf("%w: пост с id %s", customerrors.ErrAlreadyExists, post.ID)
	}

	post.CreatedAt = time.Now()
	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) UpdatePost(post *models.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.posts[post.ID]
	if !ok {
		return fmt.Errorf("%w: пост с id %s", customerrors.ErrNotFound, post.ID)
	}
	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) CreateComment(comment *models.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.comments[comment.ID]
	if ok {
		return fmt.Errorf("%w: comment with id %s", customerrors.ErrAlreadyExists, comment.ID)

	}
	comment.CreatedAt = time.Now()
	m.comments[comment.ID] = comment
	return nil
}

func (m *MemoryStorage) GetCommentByID(id string) (*models.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	comment, ok := m.comments[id]
	if !ok {
		return nil, fmt.Errorf("%w: comment with id %s", customerrors.ErrNotFound, id)
	}
	return comment, nil
}

func (m *MemoryStorage) ListCommentsByPost(postID string, parentID *string, offset, limit int) ([]*models.Comment, error) {
	if offset < 0 || limit <= 0 {
		return nil, fmt.Errorf("%w: неправильный параметр пагинации", customerrors.ErrParamOutOfRange)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	res := []*models.Comment{}
	for _, val := range m.comments {
		if val.PostID == postID {
			if parentID == nil && val.ParentID == nil {
				res = append(res, val)
			} else if parentID != nil && val.ParentID != nil && *parentID == *val.ParentID {
				res = append(res, val)
			}
		}
	}

	if offset > len(res) {
		return []*models.Comment{}, nil
	}

	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	return res[offset:end], nil
}
