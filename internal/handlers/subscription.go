package handlers

import (
	"log"

	"github.com/Harshal292004/subscription-service/internal/middleware"
	"github.com/Harshal292004/subscription-service/internal/services"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type SubscriptionHandler struct {
	service *services.SubscriptionService
}

type UserIdInput struct {
	UserId int `json:"userId"`
}

type PlanIdInput struct {
	PlanId int `json:"planId"`
}

func NewSubscriptionHandler(s *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: s,
	}
}

// RegisterSubscriptionRoutes godoc
// @Summary     Manage user subscriptions
// @Tags        subscriptions
// @Security    BearerAuth
func RegisterSubscriptionRoutes(r fiber.Router, service *services.SubscriptionService) {
	log.Println("[RegisterSubscriptionRoutes] Registering subscription routes")
	h := &SubscriptionHandler{service}
	r.Use(middleware.AuthMiddleware())

	r.Get("/subscription", h.GetSubscription)
	r.Post("/subscription", h.PostSubscription)
	r.Delete("/subscription", h.DeleteSubscription)
	r.Put("/subscription", h.PutSubscription)
	log.Println("[RegisterSubscriptionRoutes] All subscription routes registered successfully")
}

// GetSubscription godoc
// @Summary     Get current subscription for a user
// @Description Get subscription for authenticated user
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /api/subs/subscription [get]
// @Security    BearerAuth
func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	log.Println("[GetSubscription] === Starting GetSubscription request ===")

	// Get userID from middleware context with better type checking
	userIDRaw := c.Locals("userId")
	log.Printf("[GetSubscription] Raw userID from context: %v (type: %T)", userIDRaw, userIDRaw)

	var userID int
	var ok bool

	// Try int first (our expected type)
	if userID, ok = userIDRaw.(int); !ok {
		// Try uint as fallback (in case middleware stored as uint)
		if userIDUint, okUint := userIDRaw.(uint); okUint {
			userID = int(userIDUint)
			log.Printf("[GetSubscription] Converted uint to int: %d", userID)
		} else {
			log.Printf("[GetSubscription] Failed to extract userID from context. Raw value: %v, Type: %T", userIDRaw, userIDRaw)
			log.Println("[GetSubscription] === Returning 400 error ===")
			return c.Status(400).JSON(fiber.Map{"error": "Invalid user context"})
		}
	}

	log.Printf("[GetSubscription] Successfully extracted userID: %d", userID)
	log.Printf("[GetSubscription] Calling service.GetSubscription for userID: %d", userID)

	sub, err := h.service.GetSubscription(userID)
	if err != nil {
		log.Printf("[GetSubscription] Service returned error: %v", err)
		log.Println("[GetSubscription] === Returning 500 error ===")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("[GetSubscription] Successfully retrieved subscription for userID: %d", userID)
	log.Println("[GetSubscription] === Returning successful response ===")
	return c.JSON(fiber.Map{"data": sub})
}

// PostSubscription godoc
// @Summary     Create a new subscription
// @Description Provide planId in request body to subscribe
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param       input body PlanIdInput true "Plan ID input"
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router		/api/subs/subscription [post]
// @Security    BearerAuth
func (h *SubscriptionHandler) PostSubscription(c *fiber.Ctx) error {
	log.Println("[PostSubscription] === Starting PostSubscription request ===")

	// Get userID from middleware context
	userID, ok := c.Locals("userId").(int)
	if !ok {
		log.Println("[PostSubscription] Failed to extract userID from context")
		log.Println("[PostSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user context"})
	}

	log.Printf("[PostSubscription] Extracted userID from context: %d", userID)

	// Parse plan ID from request body
	var planInput PlanIdInput
	log.Println("[PostSubscription] Parsing request body for planId")
	if err := c.BodyParser(&planInput); err != nil {
		log.Printf("[PostSubscription] Failed to parse request body: %v", err)
		log.Println("[PostSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Printf("[PostSubscription] Successfully parsed planId from body: %d", planInput.PlanId)
	log.Printf("[PostSubscription] Validating input struct for planId: %d", planInput.PlanId)

	if err := utils.ValidateStruct(planInput); err != nil {
		log.Printf("[PostSubscription] Struct validation failed: %v", err)
		log.Println("[PostSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("[PostSubscription] Input validation successful")
	log.Printf("[PostSubscription] Calling service.PostSubscription for userID: %d, planId: %d", userID, planInput.PlanId)

	sub, err := h.service.PostSubscription(userID, planInput.PlanId)
	if err != nil {
		log.Printf("[PostSubscription] Service returned error: %v", err)
		log.Println("[PostSubscription] === Returning 500 error ===")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("[PostSubscription] Successfully created subscription for userID: %d, planId: %d", userID, planInput.PlanId)
	log.Println("[PostSubscription] === Returning successful response ===")
	return c.JSON(fiber.Map{"data": sub})
}

// DeleteSubscription godoc
// @Summary     Delete user subscription
// @Description Delete subscription for authenticated user
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /api/subs/subscription [delete]
// @Security    BearerAuth
func (h *SubscriptionHandler) DeleteSubscription(c *fiber.Ctx) error {
	log.Println("[DeleteSubscription] === Starting DeleteSubscription request ===")

	// Get userID from middleware context
	userID, ok := c.Locals("userId").(int)
	if !ok {
		log.Println("[DeleteSubscription] Failed to extract userID from context")
		log.Println("[DeleteSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user context"})
	}

	log.Printf("[DeleteSubscription] Extracted userID from context: %d", userID)
	log.Printf("[DeleteSubscription] Calling service.DeleteSubscription for userID: %d", userID)

	sub, err := h.service.DeleteSubscription(userID)
	if err != nil {
		log.Printf("[DeleteSubscription] Service returned error: %v", err)
		log.Println("[DeleteSubscription] === Returning 500 error ===")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("[DeleteSubscription] Successfully deleted subscription for userID: %d", userID)
	log.Println("[DeleteSubscription] === Returning successful response ===")
	return c.JSON(fiber.Map{"data": sub})
}

// PutSubscription godoc
// @Summary     Update subscription plan for a user
// @Description Provide newPlanId in request body to update subscription
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param       input body PlanIdInput true "New plan ID input"
// @Success     200 {object} models.Subscription
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /api/subs/subscription [put]
// @Security    BearerAuth
func (h *SubscriptionHandler) PutSubscription(c *fiber.Ctx) error {
	log.Println("[PutSubscription] === Starting PutSubscription request ===")

	// Get userID from middleware context
	userID, ok := c.Locals("userId").(int)
	if !ok {
		log.Println("[PutSubscription] Failed to extract userID from context")
		log.Println("[PutSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user context"})
	}

	log.Printf("[PutSubscription] Extracted userID from context: %d", userID)

	// Parse plan ID from request body
	var planInput PlanIdInput
	log.Println("[PutSubscription] Parsing request body for new planId")
	if err := c.BodyParser(&planInput); err != nil {
		log.Printf("[PutSubscription] Failed to parse request body: %v", err)
		log.Println("[PutSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Printf("[PutSubscription] Successfully parsed new planId from body: %d", planInput.PlanId)
	log.Printf("[PutSubscription] Validating input struct for planId: %d", planInput.PlanId)

	if err := utils.ValidateStruct(planInput); err != nil {
		log.Printf("[PutSubscription] Struct validation failed: %v", err)
		log.Println("[PutSubscription] === Returning 400 error ===")
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("[PutSubscription] Input validation successful")
	log.Printf("[PutSubscription] Calling service.PutSubscription for userID: %d, new planId: %d", userID, planInput.PlanId)

	sub, err := h.service.PutSubscription(userID, planInput.PlanId)
	if err != nil {
		log.Printf("[PutSubscription] Service returned error: %v", err)
		log.Println("[PutSubscription] === Returning 500 error ===")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Printf("[PutSubscription] Successfully updated subscription for userID: %d, new planId: %d", userID, planInput.PlanId)
	log.Println("[PutSubscription] === Returning successful response ===")
	return c.JSON(fiber.Map{"data": sub})
}
