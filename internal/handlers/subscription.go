package handlers

import (
	"github.com/Harshal292004/subscription-service/internal/middleware"
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type SubscriptionHandler struct {
	service *services.SubscriptionService
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

// RegisterSubscriptionRoutes godoc
// @Summary     Manage user subscriptions
// @Tags        subscriptions
// @Security    BearerAuth
func RegisterSubscriptionRoutes(r fiber.Router, service *services.SubscriptionService) {
	h := &SubscriptionHandler{service}
	r.Use(middleware.AuthMiddleware())

	r.Get("/subscription", h.GetSubscription)
	r.Post("/subscription", h.PostSubscription)
	r.Delete("/subscription", h.DeleteSubscription)
	r.Put("/subscription", h.PutSubscription)
}

// GetSubscription godoc
// @Summary     Get current subscription for a user
// @Description Provide userId in request body to get subscription
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param       input body GetDeleteInput true "User ID input"
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /subscription [get]
// @Security    BearerAuth
func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	var input GetDeleteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	sub, err := h.service.GetSubscription(input.UserId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": sub})
}

// PostSubscription godoc
// @Summary     Create a new subscription
// @Description Provide userId and planId in request body to subscribe
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param       input body PostInput true "Subscription creation input"
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /subscription [post]
// @Security    BearerAuth
func (h *SubscriptionHandler) PostSubscription(c *fiber.Ctx) error {
	var input PostInput
	if err := c.BodyParser(&input); err != nil {
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

// DeleteSubscription godoc
// @Summary     Delete user subscription
// @Description Provide userId in request body to delete subscription
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param       input body GetDeleteInput true "User ID input"
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /subscription [delete]
// @Security    BearerAuth
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

// PutSubscription godoc
// @Summary     Update subscription plan for a user
// @Description Provide userId and newPlanId in request body to update subscription
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param       input body PutInput true "Subscription update input"
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /subscription [put]
// @Security    BearerAuth
func (h *SubscriptionHandler) PutSubscription(c *fiber.Ctx) error {
	var input PutInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if err := utils.ValidateStruct(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	sub, err := h.service.PutSubscription(input.UserId, input.NewPlanId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": sub})
}
