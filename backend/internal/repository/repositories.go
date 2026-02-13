package repository

import "github.com/apk471/go-taskmanager/internal/server"

type Repositories struct{
	ToDo *TodoRepository
	Comment *CommentRepository
	Category *CategoryRepository

}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		ToDo: NewTodoRepository(s),
		Comment: NewCommentRepository(s),
		Category: NewCategoryRepository(s),
	}
}