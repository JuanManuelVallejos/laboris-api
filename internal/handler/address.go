package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/geocoding"
	"github.com/laboris/laboris-api/internal/usecase"
)

type AddressHandler struct {
	uc *usecase.AddressUseCase
}

func NewAddressHandler(uc *usecase.AddressUseCase) *AddressHandler {
	return &AddressHandler{uc: uc}
}

func addressErrStatus(err error) int {
	if errors.Is(err, geocoding.ErrNoResults) {
		return http.StatusBadRequest
	}
	if errors.Is(err, usecase.ErrUserNotOnboarded) || errors.Is(err, usecase.ErrAddressNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, usecase.ErrAddressInUse) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func (h *AddressHandler) List(c *gin.Context) {
	clerkID := c.GetString("userId")
	addrs, err := h.uc.List(clerkID)
	if err != nil {
		c.JSON(addressErrStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, addrs)
}

type addressBody struct {
	Label   string `json:"label"   binding:"required"`
	Address string `json:"address" binding:"required"`
}

func (h *AddressHandler) Create(c *gin.Context) {
	var req addressBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clerkID := c.GetString("userId")
	addr, err := h.uc.Create(clerkID, req.Label, req.Address)
	if err != nil {
		c.JSON(addressErrStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, addr)
}

func (h *AddressHandler) Update(c *gin.Context) {
	var req addressBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clerkID := c.GetString("userId")
	addr, err := h.uc.Update(clerkID, c.Param("id"), req.Label, req.Address)
	if err != nil {
		c.JSON(addressErrStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, addr)
}

func (h *AddressHandler) Delete(c *gin.Context) {
	clerkID := c.GetString("userId")
	if err := h.uc.Delete(clerkID, c.Param("id")); err != nil {
		c.JSON(addressErrStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AddressHandler) SetDefault(c *gin.Context) {
	clerkID := c.GetString("userId")
	if err := h.uc.SetDefault(clerkID, c.Param("id")); err != nil {
		c.JSON(addressErrStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
