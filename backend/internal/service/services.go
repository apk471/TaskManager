package service

import (
	job "github.com/apk471/go-taskmanager/internal/lib/jobs"
	"github.com/apk471/go-taskmanager/internal/repository"
	"github.com/apk471/go-taskmanager/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
	ToDo *TodoService
	Comment *CommentService
	Category *CategoryService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	todoService := NewTodoService(s , repos.ToDo , repos.Category)
	commentService := NewCommentService(s , repos.Comment , repos.ToDo)
	categoryService := NewCategoryService(s , repos.Category)


	return &Services{
		Job:  s.Job,
		Auth: authService,
		ToDo: todoService,
		Comment: commentService,
		Category: categoryService,
	}, nil
}