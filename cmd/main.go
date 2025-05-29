package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Harshal292004/subscription-service/internal/config"
	"github.com/Harshal292004/subscription-service/internal/handlers"
	"github.com/Harshal292004/subscription-service/internal/middleware"
	"github.com/Harshal292004/subscription-service/internal/repository"
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	// Swagger handler
	_ "github.com/Harshal292004/subscription-service/docs"
	swagger "github.com/swaggo/fiber-swagger"
)

func main() {
	utils.InitLogger()

	logrus.Info("Starting subscription service")

	// Load env and config
	config.LoadEnv()

	// Initialize dependencies
	db, err := config.InitPostgres()
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to PostgreSQL")
	}

	redisClient := config.InitRedis()
	defer func(r *redis.Client) {
		if err := r.Close(); err != nil {
			logrus.WithError(err).Warn("failed to close Redis")
		}
	}(redisClient)

	repo := repository.NewRepository(db, redisClient)

	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		AppName:       "SubscriptionService",
	})

	// Middleware
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	// Swagger endpoint
	app.Get("/swagger/*", swagger.WrapHandler)

	// Route registration
	registerRoutes(app, repo)

	go func() {
		if err := app.Listen(":8080"); err != nil {
			logrus.WithError(err).Fatal("Fiber app failed")
		}
	}()

	// Start Cron Job
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go config.StartCronJobs(ctx, services.NewSubscriptionService(repo))

	// Graceful shutdown
	gracefulShutdown(app, cancel, db)

}

func registerRoutes(app *fiber.App, repo *repository.Repository) {
	api := app.Group("/api")

	userService := services.NewUserService(repo)
	planService := services.NewPlanService(repo)
	subService := services.NewSubscriptionService(repo)

	handlers.RegisterUserRoutes(api.Group("/user"), userService)
	handlers.RegisterPlanRoutes(api.Group("/plans"), planService)
	handlers.RegisterSubscriptionRoutes(api.Group("/subs"), subService)
}
func gracefulShutdown(app *fiber.App, cancel context.CancelFunc, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Info("Shutting down server...")

	cancel() // stop background cron jobs

	if err := app.Shutdown(); err != nil {
		logrus.WithError(err).Fatal("failed to shutdown server cleanly")
	}

	// Close DB
	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		logrus.WithError(err).Warn("failed to close PostgreSQL connection")
	}

	logrus.Info("Server stopped")
}
