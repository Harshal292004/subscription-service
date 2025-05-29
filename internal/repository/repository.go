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

func (r *Repository) GetCachedPlans() ([]models.Plan, error) {
	ctx := context.Background()
	key := "plans"
	// Try fetching from Redis
	val, err := r.Redis.Get(ctx, key).Result()
	if err == nil {
		var plans []models.Plan
		if err := json.Unmarshal([]byte(val), &plans); err == nil {
			return plans, nil
		}
	}

	// If unmarshal failed or not in cache fetch from db
	var plans []models.Plan
	if err := r.DB.WithContext(ctx).Find(&plans).Error; err != nil {
		return nil, err
	}

	// Marshal the plans fetched and cache
	data, err := json.Marshal(plans)
	if err == nil {
		r.Redis.Set(ctx, key, data, 0)
	}

	return plans, nil
}

func (r *Repository) GetCachedSubscription(userId int) (models.Subscription, error) {
	ctx := context.Background()
	// Key is userid:sub (eg. (1:sub,sub{}))
	key := fmt.Sprintf("%s:sub", fmt.Sprint(userId))

	// Try fetching from Redis
	// gorm understand the table to query by conveting Subscription to subscriptions or by the data type being passed to query
	val, err := r.Redis.Get(ctx, key).Result()
	if err == nil {
		var sub models.Subscription
		if err := json.Unmarshal([]byte(val), &sub); err == nil {
			return sub, nil
		}
	}

	// If unmarshal fails or user didn't subscribe
	// You need to throw the correct error so checking whether data base has the entire for the user subscription
	var sub models.Subscription
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	// Marshal and chache sub in redis
	// This is aggresive caching as sub will also be cached when created
	data, err := json.Marshal(sub)
	if err != nil {
		return sub, nil
	}

	// Set ttl based on EndDate
	ttl := time.Until(sub.EndDate)
	if ttl <= 0 {
		ttl = time.Hour // Default to small TTL if already expired or invalid
	}
	// Set the TTL to the number of days to expire
	r.Redis.Set(ctx, key, data, ttl)

	return sub, nil
}

func (r *Repository) PostSubscription(userId int, planId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:sub", fmt.Sprint(userId))

	// Fetch the plan with the planId
	// Return error if plan invalid
	var plan models.Plan
	if err := r.DB.WithContext(ctx).First(&plan, planId).Error; err != nil {
		return models.Subscription{}, err
	}

	// Get plan duration and setup the Subscription accordingly
	now := time.Now()
	endDate := now.AddDate(0, 0, plan.Duration)

	sub := models.Subscription{
		UserID:    uint(userId),
		PlanID:    uint(planId),
		Status:    models.Active,
		StartDate: now,
		EndDate:   endDate,
	}

	// save to DB
	if err := r.DB.WithContext(ctx).Create(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	// Cache in Redis
	data, err := json.Marshal(sub)
	if err == nil {
		ttl := time.Until(endDate)
		if ttl <= 0 {
			ttl = time.Hour
		}
		r.Redis.Set(ctx, key, data, ttl)
	}

	return sub, nil
}

func (r *Repository) DeleteSubscription(userId uint) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:sub", fmt.Sprint(userId))

	// If no subscription found
	var sub models.Subscription
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	// Update the status of the subscription to cancelled (here we will just delete for simplicity but could be used for analysis)
	sub.Status = models.Cancelled

	// Mock updation of data base
	if err := r.DB.WithContext(ctx).Save(&sub).Error; err != nil {
		return models.Subscription{}, err
	}
	// We will just delete the sub for now
	if err := r.DB.WithContext(ctx).Delete(&models.Subscription{}, sub.ID).Error; err != nil {
		return models.Subscription{}, err
	}

	// Find Sub in cache and invalidate it
	r.Redis.Del(ctx, key)
	// Send the delete sub as a response
	return sub, nil
}

func (r *Repository) PutSubscription(userId int, newPlanId int) (models.Subscription, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:sub", fmt.Sprint(userId))

	// Check for existing subscription
	var sub models.Subscription
	if err := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error; err != nil {
		return models.Subscription{}, err
	}

	// Fetch new plan
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
		if ttl <= 0 {
			ttl = time.Hour
		}
		r.Redis.Set(ctx, key, data, ttl)
	}

	return sub, nil
}

func (r *Repository) PostUser(name string, password string) (string, error) {
	ctx := context.Background()
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	var jwt_secret = os.Getenv("JWT_SECRET")

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
	signedToken, err := token.SignedString(jwt_secret)
	if err != nil {
		return "", err
	}

	sessionKey := fmt.Sprintf("user:%d:session", user.ID)
	if err := r.Redis.Set(ctx, sessionKey, signedToken, 24*time.Hour).Err(); err != nil {
		return "", err
	}

	return signedToken, nil
}
