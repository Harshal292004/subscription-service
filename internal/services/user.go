package services

import "github.com/Harshal292004/subscription-service/internal/repository"

type UserService struct {
	repo *repository.Repository
}

func NewUserService(r *repository.Repository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) RegisterUser(name, password string) (string, error) {
	return s.repo.PostUser(name, password)
}
