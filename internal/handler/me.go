package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type MeHandler struct {
	uc *usecase.MeUseCase
}

func NewMeHandler(uc *usecase.MeUseCase) *MeHandler {
	return &MeHandler{uc: uc}
}

func (h *MeHandler) GetMyProfessional(c *gin.Context) {
	clerkID := c.GetString("userId")
	prof, err := h.uc.GetMyProfessional(clerkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if prof == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "professional profile not found"})
		return
	}
	c.JSON(http.StatusOK, prof)
}
