package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
	"github.com/laboris/laboris-api/internal/webhook"
)

type ClerkWebhookHandler struct {
	uc     *usecase.UserSyncUseCase
	secret string
}

func NewClerkWebhookHandler(uc *usecase.UserSyncUseCase, secret string) *ClerkWebhookHandler {
	return &ClerkWebhookHandler{uc: uc, secret: secret}
}

type clerkWebhookEvent struct {
	Type string `json:"type"`
	Data struct {
		ID       string `json:"id"`
		ImageURL string `json:"image_url"`
	} `json:"data"`
}

// Handle procesa los eventos user.deleted / user.updated que Clerk manda por
// webhook (firmados con el esquema de Svix), para mantener la app en línea
// con acciones que el usuario hace directamente en Clerk.
func (h *ClerkWebhookHandler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo leer el body"})
		return
	}

	err = webhook.Verify(
		h.secret,
		c.GetHeader("svix-id"),
		c.GetHeader("svix-timestamp"),
		c.GetHeader("svix-signature"),
		body,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var event clerkWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido"})
		return
	}

	var syncErr error
	switch event.Type {
	case "user.deleted":
		syncErr = h.uc.SoftDelete(event.Data.ID)
	case "user.updated":
		syncErr = h.uc.UpdateAvatarURL(event.Data.ID, event.Data.ImageURL)
	}
	if syncErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": syncErr.Error()})
		return
	}

	c.Status(http.StatusOK)
}
