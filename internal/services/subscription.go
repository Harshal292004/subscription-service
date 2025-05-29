package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Harshal292004/subscription-service/internal/models"
	"github.com/Harshal292004/subscription-service/internal/repository"
)

type SubscriptionService struct {
	repo *repository.Repository
}

func NewSubscriptionService(r *repository.Repository) *SubscriptionService {
	return &SubscriptionService{repo: r}
}

func (s *SubscriptionService) GetSubscription(userId int) (models.Subscription, error) {
	return s.repo.GetCachedSubscription(userId)
}

func (s *SubscriptionService) PostSubscription(userId int, planId int) (models.Subscription, error) {
	return s.repo.PostSubscription(userId, planId)
}

func (s *SubscriptionService) DeleteSubscription(userId int) (models.Subscription, error) {
	return s.repo.DeleteSubscription(userId)
}

func (s *SubscriptionService) PutSubscription(userId int, newPlanId int) (models.Subscription, error) {
	return s.repo.PutSubscription(userId, newPlanId)
}

func (s *SubscriptionService) CheckExpiredSubscriptions() error {
	now := time.Now()
	var expiredSubs []models.Subscription

	err := s.repo.DB.
		Where("end_date <= ? AND status = ?", now, models.Active).
		Find(&expiredSubs).Error
	if err != nil {
		return err
	}

	for _, sub := range expiredSubs {
		sub.Status = models.Expired
		s.repo.DB.Save(&sub)
		s.repo.Redis.Del(context.Background(), fmt.Sprintf("%d:sub", sub.UserID))
	}

	return nil
}
