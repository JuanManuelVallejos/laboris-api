package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type RequestHandler struct {
	uc *usecase.RequestUseCase
}

func NewRequestHandler(uc *usecase.RequestUseCase) *RequestHandler {
	return &RequestHandler{uc: uc}
}

type createRequestBody struct {
	ProfessionalID string `json:"professionalId" binding:"required"`
	Description    string `json:"description"    binding:"required"`
}

func (h *RequestHandler) Create(c *gin.Context) {
	var req createRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clerkID := c.GetString("userId")
	r, err := h.uc.Create(clerkID, req.ProfessionalID, req.Description)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrSelfRequest) || errors.Is(err, usecase.ErrAddressRequired) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *RequestHandler) ListReceived(c *gin.Context) {
	clerkID := c.GetString("userId")
	requests, err := h.uc.ListReceivedByProfessional(clerkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *RequestHandler) ListSent(c *gin.Context) {
	clerkID := c.GetString("userId")
	requests, err := h.uc.ListSentByClient(clerkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *RequestHandler) GetReceivedDetail(c *gin.Context) {
	clerkID := c.GetString("userId")
	r, err := h.uc.GetReceivedRequestDetail(clerkID, c.Param("id"))
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

type updateStatusBody struct {
	Status string `json:"status"          binding:"required,oneof=accepted rejected"`
	Reason string `json:"rejectionReason"`
}

func (h *RequestHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clerkID := c.GetString("userId")
	r, err := h.uc.UpdateStatus(clerkID, id, req.Status, req.Reason)
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}
