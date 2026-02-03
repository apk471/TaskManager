package service

import (
	job "github.com/apk471/go-taskmanager/internal/lib/jobs"
	"github.com/apk471/go-taskmanager/internal/repository"
	"github.com/apk471/go-taskmanager/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}