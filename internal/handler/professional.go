package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type ProfessionalHandler struct {
	uc usecase.ProfessionalUseCase
}

func NewProfessionalHandler(uc usecase.ProfessionalUseCase) *ProfessionalHandler {
	return &ProfessionalHandler{uc: uc}
}

func (h *ProfessionalHandler) GetAll(c *gin.Context) {
	clerkID := c.GetString("userId")
	professionals, err := h.uc.GetAll(clerkID)
	if err != nil {
		if errors.Is(err, usecase.ErrAddressRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, professionals)
}

func (h *ProfessionalHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	professional, err := h.uc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "professional not found"})
		return
	}
	c.JSON(http.StatusOK, professional)
}
