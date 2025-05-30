package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	log.Println("[NewRepository] Creating new repository instance")
	return &Repository{
		DB:    db,
		Redis: redis,
	}
}

func (r *Repository) GetCachedPlans() ([]models.Plan, error) {
	log.Println("[GetCachedPlans] === Starting GetCachedPlans ===")
	ctx := context.Background()
	key := "plans"

	var val string
	log.Printf("[GetCachedPlans] Attempting to get cached plans from Redis with key: %s", key)

	err := retry.Do(func() error {
		var err error
		val, err = r.Redis.Get(ctx, key).Result()
		if err != nil {
			log.Printf("[GetCachedPlans] Redis GET attempt failed: %v", err)
		} else {
			log.Printf("[GetCachedPlans] Redis GET successful, data length: %d", len(val))
		}
		return err
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err == nil {
		log.Println("[GetCachedPlans] Cache hit - attempting to unmarshal data")
		var plans []models.Plan
		if err := json.Unmarshal([]byte(val), &plans); err == nil {
			log.Printf("[GetCachedPlans] Successfully unmarshaled %d plans from cache", len(plans))
			log.Println("[GetCachedPlans] === Returning cached plans ===")
			return plans, nil
		} else {
			log.Printf("[GetCachedPlans] Failed to unmarshal Redis data: %v", err)
			log.Printf("[GetCachedPlans] Raw Redis data: %s", val)
		}
	} else {
		log.Printf("[GetCachedPlans] Redis cache miss or error: %v", err)
	}

	log.Println("[GetCachedPlans] Falling back to database query")
	var plans []models.Plan

	err = retry.Do(func() error {
		log.Println("[GetCachedPlans] Attempting DB query")
		dbErr := r.DB.WithContext(ctx).Find(&plans).Error
		if dbErr != nil {
			log.Printf("[GetCachedPlans] DB query attempt failed: %v", dbErr)
		} else {
			log.Printf("[GetCachedPlans] DB query successful, found %d plans", len(plans))
		}
		return dbErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[GetCachedPlans] All DB query attempts failed: %v", err)
		log.Println("[GetCachedPlans] === Returning error ===")
		return nil, err
	}

	log.Printf("[GetCachedPlans] Successfully fetched %d plans from DB", len(plans))

	// Cache the results
	data, marshalErr := json.Marshal(plans)
	if marshalErr == nil {
		log.Printf("[GetCachedPlans] Successfully marshaled plans data, size: %d bytes", len(data))
		log.Println("[GetCachedPlans] Attempting to cache plans in Redis")

		cacheErr := retry.Do(func() error {
			setCacheErr := r.Redis.Set(ctx, key, data, 12*time.Hour).Err()
			if setCacheErr != nil {
				log.Printf("[GetCachedPlans] Redis SET attempt failed: %v", setCacheErr)
			} else {
				log.Println("[GetCachedPlans] Redis SET successful")
			}
			return setCacheErr
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

		if cacheErr != nil {
			log.Printf("[GetCachedPlans] Failed to cache data in Redis after retries: %v", cacheErr)
		} else {
			log.Println("[GetCachedPlans] Successfully cached plans in Redis")
		}
	} else {
		log.Printf("[GetCachedPlans] Failed to marshal DB data for caching: %v", marshalErr)
	}

	log.Println("[GetCachedPlans] === Returning DB plans ===")
	return plans, nil
}

func (r *Repository) GetCachedSubscription(userId int) (models.Subscription, error) {
	log.Printf("[GetCachedSubscription] === Starting GetCachedSubscription for user ID: %d ===", userId)
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)
	log.Printf("[GetCachedSubscription] Using Redis key: %s", key)

	var val string
	err := retry.Do(func() error {
		var err error
		val, err = r.Redis.Get(ctx, key).Result()
		if err != nil {
			log.Printf("[GetCachedSubscription] Redis GET attempt failed: %v", err)
		} else {
			log.Printf("[GetCachedSubscription] Redis GET successful, data length: %d", len(val))
		}
		return err
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err == nil {
		log.Println("[GetCachedSubscription] Cache hit - attempting to unmarshal subscription")
		var sub models.Subscription
		if unmarshalErr := json.Unmarshal([]byte(val), &sub); unmarshalErr == nil {
			log.Printf("[GetCachedSubscription] Successfully unmarshaled subscription: ID=%d, Status=%v", sub.ID, sub.Status)
			log.Println("[GetCachedSubscription] === Returning cached subscription ===")
			return sub, nil
		} else {
			log.Printf("[GetCachedSubscription] Failed to unmarshal cached data: %v", unmarshalErr)
			log.Printf("[GetCachedSubscription] Raw cached data: %s", val)
		}
	} else {
		log.Printf("[GetCachedSubscription] Cache miss or Redis error: %v", err)
	}

	log.Println("[GetCachedSubscription] Falling back to database query")
	var sub models.Subscription

	err = retry.Do(func() error {
		log.Printf("[GetCachedSubscription] Attempting DB query for user_id = %d", userId)
		dbErr := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error
		if dbErr != nil {
			log.Printf("[GetCachedSubscription] DB query attempt failed: %v", dbErr)
		} else {
			log.Printf("[GetCachedSubscription] DB query successful: ID=%d, UserID=%d, PlanID=%d, Status=%v",
				sub.ID, sub.UserID, sub.PlanID, sub.Status)
		}
		return dbErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[GetCachedSubscription] All DB query attempts failed: %v", err)
		log.Println("[GetCachedSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Cache the subscription
	data, marshalErr := json.Marshal(sub)
	if marshalErr == nil {
		ttl := time.Until(sub.EndDate)
		if ttl <= 0 {
			ttl = time.Hour
			log.Printf("[GetCachedSubscription] Subscription expired, using default TTL of 1 hour")
		}
		log.Printf("[GetCachedSubscription] Caching subscription with TTL: %v", ttl)

		cacheErr := retry.Do(func() error {
			setCacheErr := r.Redis.Set(ctx, key, data, ttl).Err()
			if setCacheErr != nil {
				log.Printf("[GetCachedSubscription] Redis SET attempt failed: %v", setCacheErr)
			} else {
				log.Println("[GetCachedSubscription] Redis SET successful")
			}
			return setCacheErr
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

		if cacheErr != nil {
			log.Printf("[GetCachedSubscription] Failed to cache subscription: %v", cacheErr)
		} else {
			log.Println("[GetCachedSubscription] Successfully cached subscription")
		}
	} else {
		log.Printf("[GetCachedSubscription] Failed to marshal subscription for caching: %v", marshalErr)
	}

	log.Println("[GetCachedSubscription] === Returning DB subscription ===")
	return sub, nil
}

func (r *Repository) PostSubscription(userId int, planId int) (models.Subscription, error) {
	log.Printf("[PostSubscription] === Starting PostSubscription for user ID: %d, plan ID: %d ===", userId, planId)
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)
	log.Printf("[PostSubscription] Will use Redis key: %s", key)

	// First, get the plan
	log.Printf("[PostSubscription] Fetching plan with ID: %d", planId)
	var plan models.Plan
	err := retry.Do(func() error {
		dbErr := r.DB.WithContext(ctx).First(&plan, planId).Error
		if dbErr != nil {
			log.Printf("[PostSubscription] Plan fetch attempt failed: %v", dbErr)
		} else {
			log.Printf("[PostSubscription] Plan fetched successfully: ID=%d, Name=%s, Duration=%d days",
				plan.ID, plan.Name, plan.Duration)
		}
		return dbErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PostSubscription] Failed to fetch plan after retries: %v", err)
		log.Println("[PostSubscription] === Returning error ===")
		return models.Subscription{}, err
	}
	log.Printf("[PostSubscription] The plan is %v", plan)
	// Create subscription
	now := time.Now()
	log.Printf("[PostSubscription] The plan duration is %v", plan.Duration)
	end := now.AddDate(0, 0, plan.Duration)
	log.Printf("[PostSubscription] Creating subscription: Start=%v, End=%v", now, end)

	sub := models.Subscription{
		UserID:    uint(userId),
		PlanID:    uint(planId),
		Status:    models.Active,
		StartDate: now,
		EndDate:   end,
	}

	log.Printf("[PostSubscription] Subscription object created: UserID=%d, PlanID=%d, Status=%v",
		sub.UserID, sub.PlanID, sub.Status)

	err = retry.Do(func() error {
		createErr := r.DB.WithContext(ctx).Create(&sub).Error
		if createErr != nil {
			log.Printf("[PostSubscription] DB create attempt failed: %v", createErr)
		} else {
			log.Printf("[PostSubscription] Subscription created successfully with ID: %d", sub.ID)
		}
		return createErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PostSubscription] Failed to create subscription after retries: %v", err)
		log.Println("[PostSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Cache the subscription
	data, marshalErr := json.Marshal(sub)
	if marshalErr == nil {
		ttl := time.Until(end)
		if ttl <= 0 {
			ttl = time.Hour
			log.Printf("[PostSubscription] Subscription expired, using default TTL of 1 hour")
		}
		log.Printf("[PostSubscription] Caching new subscription with TTL: %v", ttl)

		cacheErr := retry.Do(func() error {
			setCacheErr := r.Redis.Set(ctx, key, data, ttl).Err()
			if setCacheErr != nil {
				log.Printf("[PostSubscription] Redis SET attempt failed: %v", setCacheErr)
			} else {
				log.Println("[PostSubscription] Redis SET successful")
			}
			return setCacheErr
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

		if cacheErr != nil {
			log.Printf("[PostSubscription] Failed to cache new subscription: %v", cacheErr)
		} else {
			log.Println("[PostSubscription] Successfully cached new subscription")
		}
	} else {
		log.Printf("[PostSubscription] Failed to marshal new subscription for caching: %v", marshalErr)
	}

	log.Printf("[PostSubscription] === Successfully created subscription ID: %d ===", sub.ID)
	return sub, nil
}

func (r *Repository) DeleteSubscription(userId int) (models.Subscription, error) {
	log.Printf("[DeleteSubscription] === Starting DeleteSubscription for user ID: %d ===", userId)
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)
	log.Printf("[DeleteSubscription] Using Redis key: %s", key)

	// First, get the subscription
	log.Printf("[DeleteSubscription] Fetching existing subscription for user ID: %d", userId)
	var sub models.Subscription
	err := retry.Do(func() error {
		dbErr := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error
		if dbErr != nil {
			log.Printf("[DeleteSubscription] Subscription fetch attempt failed: %v", dbErr)
		} else {
			log.Printf("[DeleteSubscription] Found subscription: ID=%d, Status=%v", sub.ID, sub.Status)
		}
		return dbErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[DeleteSubscription] Failed to find subscription: %v", err)
		log.Println("[DeleteSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Update status to cancelled
	log.Printf("[DeleteSubscription] Updating subscription ID %d status to Cancelled", sub.ID)
	sub.Status = models.Cancelled

	err = retry.Do(func() error {
		saveErr := r.DB.WithContext(ctx).Save(&sub).Error
		if saveErr != nil {
			log.Printf("[DeleteSubscription] Save attempt failed: %v", saveErr)
		} else {
			log.Printf("[DeleteSubscription] Successfully updated subscription status to: %v", sub.Status)
		}
		return saveErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[DeleteSubscription] Failed to update subscription status: %v", err)
		log.Println("[DeleteSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Delete the subscription
	log.Printf("[DeleteSubscription] Deleting subscription ID: %d", sub.ID)
	err = retry.Do(func() error {
		deleteErr := r.DB.WithContext(ctx).Delete(&models.Subscription{}, sub.ID).Error
		if deleteErr != nil {
			log.Printf("[DeleteSubscription] Delete attempt failed: %v", deleteErr)
		} else {
			log.Printf("[DeleteSubscription] Successfully deleted subscription ID: %d", sub.ID)
		}
		return deleteErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[DeleteSubscription] Failed to delete subscription: %v", err)
		log.Println("[DeleteSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Remove from cache
	log.Printf("[DeleteSubscription] Removing subscription from Redis cache with key: %s", key)
	cacheErr := retry.Do(func() error {
		delErr := r.Redis.Del(ctx, key).Err()
		if delErr != nil {
			log.Printf("[DeleteSubscription] Redis DEL attempt failed: %v", delErr)
		} else {
			log.Println("[DeleteSubscription] Successfully removed from Redis cache")
		}
		return delErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if cacheErr != nil {
		log.Printf("[DeleteSubscription] Failed to remove from cache (non-critical): %v", cacheErr)
	}

	log.Printf("[DeleteSubscription] === Successfully deleted subscription for user ID: %d ===", userId)
	return sub, nil
}

func (r *Repository) PutSubscription(userId int, newPlanId int) (models.Subscription, error) {
	log.Printf("[PutSubscription] === Starting PutSubscription for user ID: %d, new plan ID: %d ===", userId, newPlanId)
	ctx := context.Background()
	key := fmt.Sprintf("%d:sub", userId)
	log.Printf("[PutSubscription] Using Redis key: %s", key)

	// Get existing subscription
	log.Printf("[PutSubscription] Fetching existing subscription for user ID: %d", userId)
	var sub models.Subscription
	err := retry.Do(func() error {
		dbErr := r.DB.WithContext(ctx).Where("user_id = ?", userId).First(&sub).Error
		if dbErr != nil {
			log.Printf("[PutSubscription] Subscription fetch attempt failed: %v", dbErr)
		} else {
			log.Printf("[PutSubscription] Found subscription: ID=%d, Current PlanID=%d, Status=%v",
				sub.ID, sub.PlanID, sub.Status)
		}
		return dbErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PutSubscription] Failed to find existing subscription: %v", err)
		log.Println("[PutSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Get new plan
	log.Printf("[PutSubscription] Fetching new plan with ID: %d", newPlanId)
	var newPlan models.Plan
	err = retry.Do(func() error {
		dbErr := r.DB.WithContext(ctx).First(&newPlan, newPlanId).Error
		if dbErr != nil {
			log.Printf("[PutSubscription] New plan fetch attempt failed: %v", dbErr)
		} else {
			log.Printf("[PutSubscription] New plan fetched: ID=%d, Name=%s, Duration=%d days",
				newPlan.ID, newPlan.Name, newPlan.Duration)
		}
		return dbErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PutSubscription] Failed to fetch new plan: %v", err)
		log.Println("[PutSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Update subscription
	now := time.Now()
	newEndDate := now.AddDate(0, 0, newPlan.Duration)

	log.Printf("[PutSubscription] Updating subscription: Old PlanID=%d -> New PlanID=%d", sub.PlanID, newPlanId)
	log.Printf("[PutSubscription] New dates: Start=%v, End=%v", now, newEndDate)

	sub.PlanID = uint(newPlanId)
	sub.Status = models.Active
	sub.StartDate = now
	sub.EndDate = newEndDate

	err = retry.Do(func() error {
		saveErr := r.DB.WithContext(ctx).Save(&sub).Error
		if saveErr != nil {
			log.Printf("[PutSubscription] Save attempt failed: %v", saveErr)
		} else {
			log.Printf("[PutSubscription] Successfully updated subscription: ID=%d, PlanID=%d", sub.ID, sub.PlanID)
		}
		return saveErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PutSubscription] Failed to update subscription: %v", err)
		log.Println("[PutSubscription] === Returning error ===")
		return models.Subscription{}, err
	}

	// Update cache
	data, marshalErr := json.Marshal(sub)
	if marshalErr == nil {
		ttl := time.Until(sub.EndDate)
		if ttl <= 0 {
			ttl = time.Hour
			log.Printf("[PostSubscription] Subscription expired, using default TTL of 1 hour")
		}
		log.Printf("[PutSubscription] Updating cache with new subscription data, TTL: %v", ttl)

		cacheErr := retry.Do(func() error {
			setCacheErr := r.Redis.Set(ctx, key, data, ttl).Err()
			if setCacheErr != nil {
				log.Printf("[PutSubscription] Redis SET attempt failed: %v", setCacheErr)
			} else {
				log.Println("[PutSubscription] Redis SET successful")
			}
			return setCacheErr
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

		if cacheErr != nil {
			log.Printf("[PutSubscription] Failed to update cache: %v", cacheErr)
		} else {
			log.Println("[PutSubscription] Successfully updated cache")
		}
	} else {
		log.Printf("[PutSubscription] Failed to marshal updated subscription: %v", marshalErr)
	}

	log.Printf("[PutSubscription] === Successfully updated subscription for user ID: %d ===", userId)
	return sub, nil
}

func (r *Repository) PostUser(name string, password string) (string, error) {
	log.Printf("[PostUser] === Starting PostUser for username: %s ===", name)
	ctx := context.Background()

	// Hash password
	log.Println("[PostUser] Hashing password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[PostUser] Failed to hash password: %v", err)
		log.Println("[PostUser] === Returning error ===")
		return "", err
	}
	log.Println("[PostUser] Password hashed successfully")

	// Check JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("[PostUser] JWT_SECRET not set in environment")
		log.Println("[PostUser] === Returning error ===")
		return "", fmt.Errorf("JWT_SECRET not set")
	}
	log.Printf("[PostUser] JWT_SECRET found, length: %d", len(jwtSecret))

	// Create user object
	user := models.User{
		Name:     name,
		Password: string(hashedPassword),
	}
	log.Printf("[PostUser] User object created: Name=%s", user.Name)

	// Save user to database
	log.Println("[PostUser] Attempting to save user to database")
	err = retry.Do(func() error {
		createErr := r.DB.WithContext(ctx).Create(&user).Error
		if createErr != nil {
			log.Printf("[PostUser] DB create attempt failed: %v", createErr)
		} else {
			log.Printf("[PostUser] User created successfully with ID: %d", user.ID)
		}
		return createErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PostUser] Failed to create user in DB after retries: %v", err)
		log.Println("[PostUser] === Returning error ===")
		return "", err
	}

	// Generate JWT token
	log.Printf("[PostUser] Generating JWT token for user ID: %d", user.ID)
	expirationTime := time.Now().Add(24 * time.Hour)
	log.Printf("[PostUser] Token expiration time: %v", expirationTime)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expirationTime.Unix(),
	})

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("[PostUser] Failed to sign JWT token: %v", err)
		log.Println("[PostUser] === Returning error ===")
		return "", err
	}
	log.Printf("[PostUser] JWT token generated successfully, length: %d", len(signedToken))

	// Cache session in Redis
	sessionKey := fmt.Sprintf("user:%d:session", user.ID)
	log.Printf("[PostUser] Caching session with key: %s", sessionKey)

	err = retry.Do(func() error {
		setErr := r.Redis.Set(ctx, sessionKey, signedToken, 24*time.Hour).Err()
		if setErr != nil {
			log.Printf("[PostUser] Redis session cache attempt failed: %v", setErr)
		} else {
			log.Println("[PostUser] Session cached successfully in Redis")
		}
		return setErr
	}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.DelayType(retry.BackOffDelay))

	if err != nil {
		log.Printf("[PostUser] Failed to cache JWT token in Redis after retries: %v", err)
		log.Println("[PostUser] Warning: User created but session not cached")
		// Don't return error here - user is created, just session caching failed
	} else {
		log.Printf("[PostUser] Session cached successfully for user ID: %d", user.ID)
	}

	log.Printf("[PostUser] === Successfully created user ID: %d and generated token ===", user.ID)
	return signedToken, nil
}
