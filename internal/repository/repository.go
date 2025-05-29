package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Harshal292004/subscription-service/internal/models"
	"github.com/avast/retry-go"
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

func (r *Repository) GetCachedPlans() ([]models.Plan, error) {
	ctx := context.Background()
	key := "plans"

	var val string
	err := retry.Do(func() error {
		var err error
		val, err = r.Redis.Get(ctx, key).Result()
		return err
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err == nil {
		var plans []models.Plan
		if err := json.Unmarshal([]byte(val), &plans); err == nil {
			return plans, nil
		}
	}

	var plans []models.Plan
	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Find(&plans).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(plans)
	if err == nil {
		_ = retry.Do(func() error {
			return r.Redis.Set(ctx, key, data, 12*time.Hour).Err()
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	}

	return plans, nil
}

func (r *Repository) GetCachedSubscription(userId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var val string
	err := retry.Do(func() error {
		var err error
		val, err = r.Redis.Get(ctx, key).Result()
		return err
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err == nil {
		var sub models.Subscription
		if err := json.Unmarshal([]byte(val), &sub); err == nil {
			return sub, nil
		}
	}

	var sub models.Subscription
	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(sub.EndDate)
		if ttl <= 0 {
			ttl = time.Hour
		}
		_ = retry.Do(func() error {
			return r.Redis.Set(ctx, key, data, ttl).Err()
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	}

	return sub, nil
}

func (r *Repository) PostSubscription(userId int, planId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var plan models.Plan
	err := retry.Do(func() error {
		return r.DB.WithContext(ctx).First(&plan, planId).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
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

	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Create(&sub).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(end)
		_ = retry.Do(func() error {
			return r.Redis.Set(ctx, key, data, ttl).Err()
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	}

	return sub, nil
}

func (r *Repository) DeleteSubscription(userId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var sub models.Subscription
	err := retry.Do(func() error {
		return r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	sub.Status = models.Cancelled
	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Save(&sub).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Delete(&models.Subscription{}, sub.ID).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	_ = retry.Do(func() error {
		return r.Redis.Del(ctx, key).Err()
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	return sub, nil
}

func (r *Repository) PutSubscription(userId int, newPlanId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)

	var sub models.Subscription
	err := retry.Do(func() error {
		return r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	var newPlan models.Plan
	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).First(&newPlan, newPlanId).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	now := time.Now()
	sub.PlanID = uint(newPlanId)
	sub.Status = models.Active
	sub.StartDate = now
	sub.EndDate = now.AddDate(0, 0, newPlan.Duration)

	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Save(&sub).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return models.Subscription{}, err
	}

	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(sub.EndDate)
		_ = retry.Do(func() error {
			return r.Redis.Set(ctx, key, data, ttl).Err()
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	}

	return sub, nil
}

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

	err = retry.Do(func() error {
		return r.DB.WithContext(ctx).Create(&user).Error
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
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
	err = retry.Do(func() error {
		return r.Redis.Set(ctx, sessionKey, signedToken, 24*time.Hour).Err()
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
