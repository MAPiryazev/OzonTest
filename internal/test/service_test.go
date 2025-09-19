package test

import (
	"context"
	"testing"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/models"
	"github.com/MAPiryazev/OzonTest/internal/repository/mocks"
	"github.com/MAPiryazev/OzonTest/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestService_CreateUser(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	cfg := &config.AppConfig{MinUsernameLen: 3}
	svc := service.NewService(mockForRepository, cfg)

	ctx := context.Background()
	user := &models.User{Username: "Ivan"}
	mockForRepository.EXPECT().CreateUser(ctx, gomock.Any()).Return(nil)

	if err := svc.CreateUser(ctx, user); err != nil {
		t.Fatalf("не удалось создать пользователя: %v", err)
	}
	if user.ID == "" {
		t.Fatal("у пользователя должен быть ID")
	}
}

func TestService_GetUserByID(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	svc := service.NewService(mockForRepository, &config.AppConfig{})
	ctx := context.Background()
	expectedUser := &models.User{ID: "123", Username: "Vanya"}
	mockForRepository.EXPECT().GetUserByID(ctx, "123").Return(expectedUser, nil)
	u, err := svc.GetUserByID(ctx, "123")
	if err != nil {
		t.Fatalf("не удалось получить пользователя: %v", err)
	}
	if u.Username != "Vanya" {
		t.Fatalf("ожидалось имя 'Vanya', получено '%s'", u.Username)
	}
}

func TestService_CreatePost(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	cfg := &config.AppConfig{MinUsernameLen: 3}
	svc := service.NewService(mockForRepository, cfg)
	ctx := context.Background()
	authorID := uuid.NewString()
	post := &models.Post{
		Title:    "Тестовый пост",
		Content:  " контент",
		AuthorID: authorID,
	}
	// смотрим что автор есть
	mockForRepository.EXPECT().GetUserByID(ctx, authorID).Return(&models.User{ID: authorID, Username: "Ivan"}, nil)
	mockForRepository.EXPECT().CreatePost(ctx, gomock.Any()).Return(nil)
	if err := svc.CreatePost(ctx, post); err != nil {
		t.Fatalf("не удалось создать пост: %v", err)
	}
	if post.ID == "" {
		t.Fatal("у поста должен быть ID")
	}
	if post.CreatedAt.IsZero() {
		t.Fatal("у поста должно быть время создания")
	}
}

func TestService_GetPostByID(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	svc := service.NewService(mockForRepository, &config.AppConfig{})
	ctx := context.Background()
	post := &models.Post{ID: "p1", Title: "Заголовок", Content: "текст"}
	mockForRepository.EXPECT().GetPostByID(ctx, "p1").Return(post, nil)

	got, err := svc.GetPostByID(ctx, "p1")
	if err != nil {
		t.Fatalf("не удалось получить пост: %v", err)
	}
	if got.Title != "Заголовок" {
		t.Fatalf("ожидался заголовок 'Заголовок', получено '%s'", got.Title)
	}
}

func TestService_ListPosts(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	cfg := &config.AppConfig{MaxListLimit: 10}
	svc := service.NewService(mockForRepository, cfg)
	ctx := context.Background()
	posts := []*models.Post{
		{ID: "p1", Title: "T1", Content: "C1"},
		{ID: "p2", Title: "T2", Content: "C2"}}

	mockForRepository.EXPECT().ListPosts(ctx, 0, 2).Return(posts, nil)
	list, err := svc.ListPosts(ctx, 0, 2)
	if err != nil {
		t.Fatalf("не удалось получить список постов: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ожидалось 2 поста получено %d", len(list))
	}
}

func TestService_CreateComment(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	svc := service.NewService(mockForRepository, &config.AppConfig{})
	ctx := context.Background()
	postID := uuid.NewString()
	comment := &models.Comment{
		PostID:   postID,
		Text:     "Привет",
		AuthorID: "user1",
	}
	mockForRepository.EXPECT().GetPostByID(ctx, postID).Return(&models.Post{ID: postID, CommentsEnabled: true}, nil)
	mockForRepository.EXPECT().CreateComment(ctx, gomock.Any()).Return(nil)
	if err := svc.CreateComment(ctx, comment); err != nil {
		t.Fatalf("не удалось создать комментарий: %v", err)
	}
	if comment.ID == "" {
		t.Fatal("у комментария должен быть ID")
	}
	if comment.CreatedAt.IsZero() {
		t.Fatal("у комментария должно быть  время создания")
	}
}

func TestService_GetCommentByID(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockForRepository := mocks.NewMockStorage(controller)
	svc := service.NewService(mockForRepository, &config.AppConfig{})
	ctx := context.Background()
	comment := &models.Comment{ID: "c1", Text: "Привет"}
	mockForRepository.EXPECT().GetCommentByID(ctx, "c1").Return(comment, nil)

	got, err := svc.GetCommentByID(ctx, "c1")
	if err != nil {
		t.Fatalf("не удалось получить комментарий: %v", err)
	}
	if got.Text != "Привет" {
		t.Fatalf("ожидался текст 'Привет', получено '%s'", got.Text)
	}
}
