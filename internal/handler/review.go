package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type ReviewHandler struct {
	uc *usecase.ReviewUseCase
}

func NewReviewHandler(uc *usecase.ReviewUseCase) *ReviewHandler {
	return &ReviewHandler{uc: uc}
}

type createReviewBody struct {
	Rating  int    `json:"rating"  binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *ReviewHandler) Create(c *gin.Context) {
	var req createReviewBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clerkID := c.GetString("userId")
	r, err := h.uc.Create(clerkID, c.Param("id"), req.Rating, req.Comment)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrReviewNotEligible) {
			status = http.StatusForbidden
		}
		if errors.Is(err, usecase.ErrUserNotOnboarded) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *ReviewHandler) ListByProfessional(c *gin.Context) {
	reviews, err := h.uc.ListByProfessional(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reviews)
}
