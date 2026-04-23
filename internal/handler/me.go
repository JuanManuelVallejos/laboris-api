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

type updateProfessionalRequest struct {
	Trade string `json:"trade" binding:"required"`
	Zone  string `json:"zone"  binding:"required"`
	Bio   string `json:"bio"`
}

func (h *MeHandler) UpdateMyProfessional(c *gin.Context) {
	var req updateProfessionalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clerkID := c.GetString("userId")
	prof, err := h.uc.UpdateMyProfessional(clerkID, req.Trade, req.Zone, req.Bio)
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
