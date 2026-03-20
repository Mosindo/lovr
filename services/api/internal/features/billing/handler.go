package billing

import (
	"errors"
	"io"
	"net/http"

	"example.com/api/internal/platform/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Checkout(c *gin.Context) {
	organizationID := c.GetString("organizationID")
	response, err := h.service.Checkout(c.Request.Context(), organizationID)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrOrganizationRequired):
			status = http.StatusBadRequest
			c.JSON(status, gin.H{"error": "organization context required"})
		case errors.Is(err, ErrBillingUnavailable):
			status = http.StatusServiceUnavailable
			c.JSON(status, gin.H{"error": "billing is not configured"})
		case errors.Is(err, ErrCheckoutSessionFailed):
			status = http.StatusBadGateway
			c.JSON(status, gin.H{"error": "could not create checkout session"})
		default:
			c.JSON(status, gin.H{"error": "could not start checkout"})
		}
		logger.LogHandlerError(c, "billing.checkout", status, err)
		return
	}

	c.JSON(http.StatusCreated, response)
	logger.LogHandlerEvent(c, "billing.checkout.success", http.StatusCreated, map[string]string{
		"organization_id": response.OrganizationID,
		"session_id":      response.SessionID,
	})
}

func (h *Handler) Webhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.LogHandlerError(c, "billing.webhook.read", http.StatusBadRequest, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	received, err := h.service.ProcessWebhook(c.Request.Context(), payload, c.GetHeader("Stripe-Signature"))
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrWebhookUnavailable):
			status = http.StatusServiceUnavailable
			c.JSON(status, gin.H{"error": "billing webhook is not configured"})
		case errors.Is(err, ErrInvalidWebhookSignature):
			status = http.StatusBadRequest
			c.JSON(status, gin.H{"error": "invalid stripe signature"})
		case errors.Is(err, ErrOrganizationRequired):
			status = http.StatusBadRequest
			c.JSON(status, gin.H{"error": "organization context required"})
		default:
			c.JSON(status, gin.H{"error": "could not process billing webhook"})
		}
		logger.LogHandlerError(c, "billing.webhook", status, err)
		return
	}

	c.JSON(http.StatusOK, WebhookResponse{Received: received})
}
