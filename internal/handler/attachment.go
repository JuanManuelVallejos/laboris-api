package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type AttachmentHandler struct {
	uc *usecase.AttachmentUseCase
}

func NewAttachmentHandler(uc *usecase.AttachmentUseCase) *AttachmentHandler {
	return &AttachmentHandler{uc: uc}
}

func (h *AttachmentHandler) UploadPortfolioPhoto(c *gin.Context) {
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing photo file"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	attachment, err := h.uc.UploadPortfolioPhoto(c.GetString("userId"), file, fileHeader.Filename)
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, attachment)
}

func (h *AttachmentHandler) DeletePortfolioPhoto(c *gin.Context) {
	if err := h.uc.DeletePortfolioPhoto(c.GetString("userId"), c.Param("id")); err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AttachmentHandler) UploadRequestPhoto(c *gin.Context) {
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing photo file"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	attachment, err := h.uc.UploadRequestPhoto(c.GetString("userId"), c.Param("id"), file, fileHeader.Filename)
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, attachment)
}
