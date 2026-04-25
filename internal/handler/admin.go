package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type AdminHandler struct {
	uc *usecase.AdminUseCase
}

func NewAdminHandler(uc *usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{uc: uc}
}

type paginatedResponse struct {
	Items      any   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

func parsePagination(c *gin.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	return
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, limit := parsePagination(c)
	users, total, err := h.uc.ListUsers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, paginatedResponse{
		Items:      users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	})
}

func (h *AdminHandler) ListProfessionals(c *gin.Context) {
	page, limit := parsePagination(c)
	professionals, total, err := h.uc.ListProfessionals(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, paginatedResponse{
		Items:      professionals,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	})
}

type verifyBody struct {
	Verified bool `json:"verified"`
}

func (h *AdminHandler) VerifyProfessional(c *gin.Context) {
	id := c.Param("id")
	var body verifyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.VerifyProfessional(id, body.Verified); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type statusBody struct {
	Status string `json:"status" binding:"required"`
}

func (h *AdminHandler) SetProfessionalStatus(c *gin.Context) {
	id := c.Param("id")
	var body statusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.SetProfessionalStatus(id, body.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) DeleteProfessional(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.DeleteProfessional(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
