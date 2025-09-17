package service_test

import (
	"testing"

	"github.com/MAPiryazev/OzonTest/internal/models"
	"github.com/MAPiryazev/OzonTest/internal/repository/inmemory"
	"github.com/MAPiryazev/OzonTest/internal/service"
)

func TestService_CreateUserPostComment(t *testing.T) {
	store := inmemory.NewMemoryStorage()
	svc := service.NewService(store)

	user := &models.User{Username: "test"}
	if err := svc.CreateUser(user); err != nil {
		t.Fatal(err)
	}

	post := &models.Post{
		Title:           "title",
		Content:         "content",
		AuthorID:        user.ID,
		CommentsEnabled: true,
	}
	if err := svc.CreatePost(post); err != nil {
		t.Fatal(err)
	}

	comment := &models.Comment{
		PostID:   post.ID,
		AuthorID: user.ID,
		Text:     "comment",
	}
	if err := svc.CreateComment(comment); err != nil {
		t.Fatal(err)
	}

	gotPost, _ := svc.GetPostByID(post.ID)
	if gotPost.ID != post.ID {
		t.Errorf("expected %s, got %s", post.ID, gotPost.ID)
	}
}
