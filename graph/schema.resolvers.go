package graph

import (
	"context"
	"time"

	"github.com/MAPiryazev/OzonTest/graph/model"
	internal "github.com/MAPiryazev/OzonTest/internal/models"
)

// Следующие несколько функций - слой преобразования internal моделей в graphql модели
func convertUser(user *internal.User) *model.User {
	if user == nil {
		return nil
	}
	return &model.User{ID: user.ID, Username: user.Username}
}

func convertPost(post *internal.Post) *model.Post {
	if post == nil {
		return nil
	}
	return &model.Post{
		ID:              post.ID,
		Title:           post.Title,
		Content:         post.Content,
		AuthorID:        post.AuthorID,
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       post.CreatedAt.Format(time.RFC3339),
	}
}

func convertComment(comment *internal.Comment) *model.Comment {
	if comment == nil {
		return nil
	}
	return &model.Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		AuthorID:  comment.AuthorID,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339)}
}

func convertMultPosts(posts []*internal.Post) []*model.Post {
	res := make([]*model.Post, len(posts))
	for i, val := range posts {
		res[i] = convertPost(val)
	}
	return res
}

func convertMultComments(commentArr []*internal.Comment) []*model.Comment {
	res := make([]*model.Comment, len(commentArr))
	for i, val := range commentArr {
		res[i] = convertComment(val)
	}
	return res
}

// дальше идут резолверы (в данном случае обертки над хендлерами)
func (r *mutationResolver) CreateUser(ctx context.Context, username string) (*model.User, error) {
	user, err := r.Handler.CreateUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return convertUser(user), nil
}

func (r *mutationResolver) CreatePost(ctx context.Context, title, content, authorID string, commentsEnabled bool) (*model.Post, error) {
	post, err := r.Handler.CreatePost(ctx, title, content, authorID, commentsEnabled)
	if err != nil {
		return nil, err
	}
	return convertPost(post), nil
}

func (r *mutationResolver) UpdatePost(ctx context.Context, id, title, content, userID string) (*model.Post, error) {
	post, err := r.Handler.UpdatePost(ctx, id, title, content, userID)
	if err != nil {
		return nil, err
	}
	return convertPost(post), nil
}

func (r *mutationResolver) CreateComment(ctx context.Context, postID, text, authorID string, parentID *string) (*model.Comment, error) {
	comment, err := r.Handler.CreateComment(ctx, postID, text, authorID, parentID)
	if err != nil {
		return nil, err
	}
	return convertComment(comment), nil
}

func (r *queryResolver) ListPosts(ctx context.Context, offset, limit int32) ([]*model.Post, error) {
	posts, err := r.Handler.ListPosts(ctx, int(offset), int(limit))
	if err != nil {
		return nil, err
	}
	return convertMultPosts(posts), nil
}

func (r *queryResolver) GetPost(ctx context.Context, id string) (*model.Post, error) {
	post, err := r.Handler.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertPost(post), nil
}

func (r *queryResolver) ListComments(ctx context.Context, postID string, parentID *string, offset, limit int32) ([]*model.Comment, error) {
	comments, err := r.Handler.ListComments(ctx, postID, parentID, int(offset), int(limit))
	if err != nil {
		return nil, err
	}
	return convertMultComments(comments), nil
}

// служебные, сгенерированные gqlgen
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
