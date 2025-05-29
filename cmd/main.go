package main

import (
	"log"

	"github.com/Harshal292004/subscription-service/internal/config"
	"github.com/Harshal292004/subscription-service/internal/handlers"
	"github.com/Harshal292004/subscription-service/internal/middleware"
	"github.com/Harshal292004/subscription-service/internal/repository"
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Logger middleware
	app.Use(middleware.Logger())

	// CORS
	app.Use(middleware.CORS())

	// Init DB, Redis
	db, err := config.InitPostgres()
	redis := config.InitRedis()
	if err != nil {
		panic("Error occured")
	}
	// Repos
	repo := repository.NewRepository(db, redis)

	// Services
	userService := services.NewUserService(repo)
	planService := services.NewPlanService(repo)
	subService := services.NewSubscriptionService(repo)

	// Routes
	api := app.Group("/api")

	handlers.RegisterUserRoutes(api.Group("/user"), userService)
	handlers.RegisterPlanRoutes(api.Group("/plans"), planService)
	handlers.RegisterSubscriptionRoutes(api.Group("/subs"), subService)

	log.Fatal(app.Listen(":8080"))
}
