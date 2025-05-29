package handlers

import (
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type SubscriptionHandler struct {
	service *services.SubscriptionService
}

func RegisterServicesRoutes(r fiber.Router, service *services.SubscriptionService) {
	h := &SubscriptionHandler{service}
	r.Get("/subscription", h.GetSubscription)
	r.Post("/subscription/:userId", h.PostSubscription)
	r.Delete("/subscription/:userId", h.DeleteSubscription)
	r.Put("/subscription/:userId/:planId", h.PutSubscription)
}

type GetDeleteInput struct {
	UserId int `json:"userId"`
}

type PostInput struct {
	UserId int `json:"userId"`
	PlanId int `json:"planId"`
}

type PutInput struct {
	UserId    int `json:"userId"`
	NewPlanId int `json:"newPlanId"`
}

func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	var input GetDeleteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := h.service.GetSubscription(input.UserId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}

func (h *SubscriptionHandler) PostSubscription(c *fiber.Ctx) error {
	var input PostInput
	if err := c.Locals(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	sub, err := h.service.PostSubscription(input.UserId, input.PlanId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": sub})
}

func (h *SubscriptionHandler) DeleteSubscription(c *fiber.Ctx) error {
	var input GetDeleteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	sub, err := h.service.DeleteSubscription(input.UserId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": sub})
}

func (h *SubscriptionHandler) PutSubscription(c *fiber.Ctx) error {
	var input PutInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	sub, err := h.service.PostSubscription(input.UserId, input.NewPlanId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": sub})
}
