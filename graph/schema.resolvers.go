package graph

import (
	"context"
	"time"

	"github.com/MAPiryazev/OzonTest/graph/model"
	internal "github.com/MAPiryazev/OzonTest/internal/models"
)

// Преобразование internal моделей в graphql модели
func toGraphQLUser(u *internal.User) *model.User {
	if u == nil {
		return nil
	}
	return &model.User{ID: u.ID, Username: u.Username}
}

func toGraphQLPost(p *internal.Post) *model.Post {
	if p == nil {
		return nil
	}
	return &model.Post{
		ID:              p.ID,
		Title:           p.Title,
		Content:         p.Content,
		AuthorID:        p.AuthorID,
		CommentsEnabled: p.CommentsEnabled,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphQLComment(c *internal.Comment) *model.Comment {
	if c == nil {
		return nil
	}
	return &model.Comment{
		ID:        c.ID,
		PostID:    c.PostID,
		ParentID:  c.ParentID,
		AuthorID:  c.AuthorID,
		Text:      c.Text,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphQLPosts(posts []*internal.Post) []*model.Post {
	res := make([]*model.Post, len(posts))
	for i, p := range posts {
		res[i] = toGraphQLPost(p)
	}
	return res
}

func toGraphQLComments(comments []*internal.Comment) []*model.Comment {
	res := make([]*model.Comment, len(comments))
	for i, c := range comments {
		res[i] = toGraphQLComment(c)
	}
	return res
}

// дальше идут резолверы (в данном случае обертки над хендлерами)
func (r *mutationResolver) CreateUser(ctx context.Context, username string) (*model.User, error) {
	u, err := r.Handler.CreateUser(username)
	if err != nil {
		return nil, err
	}
	return toGraphQLUser(u), nil
}

func (r *mutationResolver) CreatePost(ctx context.Context, title, content, authorID string, commentsEnabled bool) (*model.Post, error) {
	p, err := r.Handler.CreatePost(title, content, authorID, commentsEnabled)
	if err != nil {
		return nil, err
	}
	return toGraphQLPost(p), nil
}

func (r *mutationResolver) UpdatePost(ctx context.Context, id, title, content, userID string) (*model.Post, error) {
	p, err := r.Handler.UpdatePost(id, title, content, userID)
	if err != nil {
		return nil, err
	}
	return toGraphQLPost(p), nil
}

func (r *mutationResolver) CreateComment(ctx context.Context, postID, text, authorID string, parentID *string) (*model.Comment, error) {
	c, err := r.Handler.CreateComment(postID, text, authorID, parentID)
	if err != nil {
		return nil, err
	}
	return toGraphQLComment(c), nil
}

func (r *queryResolver) ListPosts(ctx context.Context, offset, limit int32) ([]*model.Post, error) {
	posts, err := r.Handler.ListPosts(int(offset), int(limit))
	if err != nil {
		return nil, err
	}
	return toGraphQLPosts(posts), nil
}

func (r *queryResolver) GetPost(ctx context.Context, id string) (*model.Post, error) {
	p, err := r.Handler.GetPost(id)
	if err != nil {
		return nil, err
	}
	return toGraphQLPost(p), nil
}

func (r *queryResolver) ListComments(ctx context.Context, postID string, parentID *string, offset, limit int32) ([]*model.Comment, error) {
	comments, err := r.Handler.ListComments(postID, parentID, int(offset), int(limit))
	if err != nil {
		return nil, err
	}
	return toGraphQLComments(comments), nil
}

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
