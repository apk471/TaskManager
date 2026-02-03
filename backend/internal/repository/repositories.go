package repository

import "github.com/apk471/go-taskmanager/internal/server"

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}