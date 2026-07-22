package service

import (
	"context"
	"time"

	"sistema-editorial/editora/backend/src/health/entity"
)

type databaseChecker interface {
	Check(ctx context.Context) string
}

type Service struct {
	appName string
	repo    databaseChecker
}

func NewService(appName string, repo databaseChecker) *Service {
	return &Service{
		appName: appName,
		repo:    repo,
	}
}

func (s *Service) Status() entity.StatusResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return entity.StatusResponse{
		Service:  s.appName,
		Status:   "ok",
		Database: s.repo.Check(ctx),
	}
}
