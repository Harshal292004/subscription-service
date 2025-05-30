package handlers

import (
	"log"

	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *services.UserService
}

// RegisterUserRoutes godoc
// @Summary     Register new users
// @Tags        users
func RegisterUserRoutes(r fiber.Router, service *services.UserService) {
	log.Println("[RegisterUserRoutes] Registering user routes")
	h := &UserHandler{service}
	r.Post("/register", h.Register)
	log.Println("[RegisterUserRoutes] User routes registered successfully")
}

type RegisterInput struct {
	Name     string `json:"name" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"`
}

// Register godoc
// @Summary     Register new user and return JWT token
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       input body RegisterInput true "User registration input"
// @Success     200 {object} map[string]string "JWT token"
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /api/user/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	log.Println("[Register] === Starting user registration request ===")

	var input RegisterInput
	log.Println("[Register] Parsing request body")
	if err := c.BodyParser(&input); err != nil {
		log.Printf("[Register] Failed to parse request body: %v", err)
		log.Println("[Register] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Printf("[Register] Successfully parsed input - Name: %s, Password length: %d", input.Name, len(input.Password))
	log.Println("[Register] Validating input struct")

	if err := utils.ValidateStruct(input); err != nil {
		log.Printf("[Register] Input validation failed: %v", err)
		log.Println("[Register] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("[Register] Input validation successful for user: %s", input.Name)
	log.Printf("[Register] Calling service.RegisterUser for: %s", input.Name)

	token, err := h.service.RegisterUser(input.Name, input.Password)
	if err != nil {
		log.Printf("[Register] Service returned error: %v", err)
		log.Println("[Register] === Returning 500 error ===")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("[Register] Successfully registered user: %s, token length: %d", input.Name, len(token))
	log.Println("[Register] === Returning successful response ===")
	return c.JSON(fiber.Map{"token": token})
}
