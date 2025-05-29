package services

import (
	"github.com/Harshal292004/subscription-service/internal/models"
	"github.com/Harshal292004/subscription-service/internal/repository"
)

type PlanService struct {
	repo *repository.Repository
}

func NewPlanService(r *repository.Repository) *PlanService {
	return &PlanService{repo: r}
}

func (s *PlanService) GetAllPlans() ([]models.Plan, error) {
	return s.repo.GetCachedPlans()
}
