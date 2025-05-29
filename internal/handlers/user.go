package handlers

import (
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *services.UserService
}

// RegisterUserRoutes godoc
// @Summary     Get all available plans
// @Tags        plans
func RegisterUserRoutes(r fiber.Router, service *services.UserService) {
	h := &UserHandler{service}
	r.Post("/register", h.Register)
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
// @Router      /register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := h.service.RegisterUser(input.Name, input.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}
