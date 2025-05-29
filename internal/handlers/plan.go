package handlers

import (
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/gofiber/fiber/v2"
)

type PlanHandler struct {
	service *services.PlanService
}

// RegisterPlanRoutes godoc
// @Summary     Get all available plans
// @Tags        plans
func RegisterPlanRoutes(r fiber.Router, service *services.PlanService) {
	h := &PlanHandler{service}
	r.Get("/plans", h.GetAllPlans)
}

// GetAllPlans godoc
// @Summary     Retrieve all plans
// @Tags        plans
// @Accept      json
// @Produce     json
// @Success     200 {array} models.Plan
// @Failure     500 {object} map[string]string
// @Router      /plans [get]
func (h *PlanHandler) GetAllPlans(c *fiber.Ctx) error {
	plans, err := h.service.GetAllPlans()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": plans})
}
