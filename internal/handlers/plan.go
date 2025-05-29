package handlers

import (
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type PlanHandler struct {
	service *services.PlanService
}

func RegisterPlanRoutes(r fiber.Router, service *services.PlanService) {
	h := &PlanHandler{service}
	r.Post("/register", h.GetAllPlans)
}

func (h *PlanHandler) GetAllPlans(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := h.service.GetAllPlans()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}
