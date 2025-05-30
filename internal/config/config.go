package config

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     string
	Password string
	User     string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Protocol int
}

func NewConection(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return db, err
	}
	return db, nil
}

func NewRedisConnection(config *RedisConfig) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		Protocol: config.Protocol,
	})

	return client
}

func InitPostgres() (*gorm.DB, error) {
	configuration := Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		User:     os.Getenv("POSTGRES_USER"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	}
	return NewConection(&configuration)
}

func InitRedis() *redis.Client {
	configuration := RedisConfig{
		Addr:     os.Getenv("ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		Protocol: 1,
	}
	return NewRedisConnection(&configuration)
}

// func StartCronJobs(ctx context.Context, subService *services.SubscriptionService) {
// 	c := cron.New()
// 	err := c.AddFunc("@every 1h", func() {
// 		log.Println("Running CheckExpiredSubscriptions job")
// 		if err := subService.CheckExpiredSubscriptions(); err != nil {
// 			log.Printf("Error running CheckExpiredSubscriptions: %v", err)
// 		}
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to schedule cron: %v", err)
// 	}
// 	c.Start()
// 	go func() {
// 		<-ctx.Done()
// 		logrus.Info("Stopping cron jobs...")
// 		c.Stop()
// 	}()
// }
