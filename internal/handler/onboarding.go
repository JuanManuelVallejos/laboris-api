package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type OnboardingHandler struct {
	uc *usecase.OnboardingUseCase
}

func NewOnboardingHandler(uc *usecase.OnboardingUseCase) *OnboardingHandler {
	return &OnboardingHandler{uc: uc}
}

type onboardingRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	FullName string `json:"fullName" binding:"required"`
	Role     string `json:"role"     binding:"required,oneof=client professional"`
	Trade    string `json:"trade"`
	Zone     string `json:"zone"`
	Bio      string `json:"bio"`
}

func (h *OnboardingHandler) Complete(c *gin.Context) {
	var req onboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "professional" && (req.Trade == "" || req.Zone == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trade and zone are required for professionals"})
		return
	}

	clerkID := c.GetString("userId")

	result, err := h.uc.Execute(usecase.OnboardingInput{
		ClerkID:  clerkID,
		Email:    req.Email,
		FullName: req.FullName,
		Role:     req.Role,
		Trade:    req.Trade,
		Zone:     req.Zone,
		Bio:      req.Bio,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "onboarding failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"userId": result.UserID, "role": result.Role})
}
