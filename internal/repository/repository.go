package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Harshal292004/subscription-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Repository struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) *Repository {
	return &Repository{
		DB:    db,
		Redis: redis,
	}
}

// GetCachedPlans attempts to fetch plans from Redis cache, falling back to DB if needed.
func (r *Repository) GetCachedPlans() ([]models.Plan, error) {
	ctx := context.Background()
	key := "plans"

	val, err := r.Redis.Get(ctx, key).Result()
	if err == nil {
		var plans []models.Plan
		if err := json.Unmarshal([]byte(val), &plans); err == nil {
			return plans, nil
		}
	}

	var plans []models.Plan
	if err := r.DB.WithContext(ctx).Find(&plans).Error; err != nil {
		return nil, err
	}

	data, err := json.Marshal(plans)
	if err == nil {
		r.Redis.Set(ctx, key, data, 12*time.Hour)
	}

	return plans, nil
}

// GetCachedSubscription fetches user subscription from cache or DB.
func (r *Repository) GetCachedSubscription(userId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	val, err := r.Redis.Get(ctx, key).Result()
	if err == nil {
		var sub models.Subscription
		if err := json.Unmarshal([]byte(val), &sub); err == nil {
			return sub, nil
		}
	}

	var sub models.Subscription
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(sub.EndDate)
		if ttl <= 0 {
			ttl = time.Hour
		}
		r.Redis.Set(ctx, key, data, ttl)
	}

	return sub, nil
}

// PostSubscription creates a new subscription for a user to a plan.
func (r *Repository) PostSubscription(userId int, planId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var plan models.Plan
	if err := r.DB.WithContext(ctx).First(&plan, planId).Error; err != nil {
		return models.Subscription{}, err
	}

	now := time.Now()
	end := now.AddDate(0, 0, plan.Duration)

	sub := models.Subscription{
		UserID:    uint(userId),
		PlanID:    uint(planId),
		Status:    models.Active,
		StartDate: now,
		EndDate:   end,
	}

	if err := r.DB.WithContext(ctx).Create(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(end)
		r.Redis.Set(ctx, key, data, ttl)
	}

	return sub, nil
}

// DeleteSubscription cancels and deletes a user's subscription.
func (r *Repository) DeleteSubscription(userId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var sub models.Subscription
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	sub.Status = models.Cancelled
	if err := r.DB.WithContext(ctx).Save(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	if err := r.DB.WithContext(ctx).Delete(&models.Subscription{}, sub.ID).Error; err != nil {
		return models.Subscription{}, err
	}

	r.Redis.Del(ctx, key)
	return sub, nil
}

// PutSubscription upgrades or downgrades a user’s subscription.
func (r *Repository) PutSubscription(userId int, newPlanId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var sub models.Subscription
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	var newPlan models.Plan
	if err := r.DB.WithContext(ctx).First(&newPlan, newPlanId).Error; err != nil {
		return models.Subscription{}, err
	}

	now := time.Now()
	sub.PlanID = uint(newPlanId)
	sub.Status = models.Active
	sub.StartDate = now
	sub.EndDate = now.AddDate(0, 0, newPlan.Duration)

	if err := r.DB.WithContext(ctx).Save(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(sub.EndDate)
		r.Redis.Set(ctx, key, data, ttl)
	}

	return sub, nil
}

// PostUser creates a new user and returns a JWT token.
func (r *Repository) PostUser(name string, password string) (string, error) {
	ctx := context.Background()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	user := models.User{
		Name:     name,
		Password: string(hashedPassword),
	}

	if err := r.DB.WithContext(ctx).Create(&user).Error; err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	sessionKey := fmt.Sprintf("user:%d:session", user.ID)
	if err := r.Redis.Set(ctx, sessionKey, signedToken, 24*time.Hour).Err(); err != nil {
		return "", err
	}

	return signedToken, nil
}
