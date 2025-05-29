package main

import (
	"log"

	"github.com/Harshal292004/subscription-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func main() {
	err := godotenv.Load(".env")
	&config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello")
	})

	app.Listen(":3000")
}
