package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/usecase"
)

type JobHandler struct {
	uc  *usecase.JobUseCase
	muc *usecase.MessageUseCase
}

func NewJobHandler(uc *usecase.JobUseCase, muc *usecase.MessageUseCase) *JobHandler {
	return &JobHandler{uc: uc, muc: muc}
}

func (h *JobHandler) GetJob(c *gin.Context) {
	clerkID := c.GetString("userId")
	job, err := h.uc.GetByID(clerkID, c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) ListMyJobs(c *gin.Context) {
	clerkID := c.GetString("userId")
	jobs, err := h.uc.ListByUser(clerkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

type scheduleVisitBody struct {
	ScheduledAt time.Time `json:"scheduledAt" binding:"required"`
}

func (h *JobHandler) ScheduleVisit(c *gin.Context) {
	var body scheduleVisitBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.uc.ScheduleVisit(c.GetString("userId"), c.Param("id"), body.ScheduledAt)
	h.respond(c, job, err)
}

type amountBody struct {
	Amount float64 `json:"amount" binding:"required"`
}

func (h *JobHandler) SubmitVisitQuote(c *gin.Context) {
	var body amountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.uc.SubmitVisitQuote(c.GetString("userId"), c.Param("id"), body.Amount)
	h.respond(c, job, err)
}

type skipVisitBody struct {
	WorkAmount      float64 `json:"workAmount"      binding:"required"`
	WorkDescription string  `json:"workDescription"`
}

func (h *JobHandler) SkipVisit(c *gin.Context) {
	var body skipVisitBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.uc.SkipVisit(c.GetString("userId"), c.Param("id"), body.WorkAmount, body.WorkDescription)
	h.respond(c, job, err)
}

func (h *JobHandler) PayVisit(c *gin.Context) {
	job, err := h.uc.PayVisit(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

func (h *JobHandler) CompleteVisit(c *gin.Context) {
	job, err := h.uc.CompleteVisit(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

type workQuoteBody struct {
	Amount      float64 `json:"amount"      binding:"required"`
	Description string  `json:"description"`
}

func (h *JobHandler) SubmitWorkQuote(c *gin.Context) {
	var body workQuoteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.uc.SubmitWorkQuote(c.GetString("userId"), c.Param("id"), body.Amount, body.Description)
	h.respond(c, job, err)
}

func (h *JobHandler) ApproveWorkQuote(c *gin.Context) {
	job, err := h.uc.ApproveWorkQuote(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

func (h *JobHandler) StartWork(c *gin.Context) {
	job, err := h.uc.StartWork(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

func (h *JobHandler) DeliverWork(c *gin.Context) {
	job, err := h.uc.DeliverWork(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

type reworkBody struct {
	Notes string `json:"notes"`
}

func (h *JobHandler) RequestRework(c *gin.Context) {
	var body reworkBody
	_ = c.ShouldBindJSON(&body)
	job, err := h.uc.RequestRework(c.GetString("userId"), c.Param("id"), body.Notes)
	h.respond(c, job, err)
}

func (h *JobHandler) SubmitReworkQuote(c *gin.Context) {
	var body amountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.uc.SubmitReworkQuote(c.GetString("userId"), c.Param("id"), body.Amount)
	h.respond(c, job, err)
}

func (h *JobHandler) ApproveReworkQuote(c *gin.Context) {
	job, err := h.uc.ApproveReworkQuote(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

func (h *JobHandler) AcceptRework(c *gin.Context) {
	job, err := h.uc.AcceptRework(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

func (h *JobHandler) ApproveDelivery(c *gin.Context) {
	job, err := h.uc.ApproveDelivery(c.GetString("userId"), c.Param("id"))
	h.respond(c, job, err)
}

type cancelBody struct {
	Reason string `json:"reason"`
}

func (h *JobHandler) Cancel(c *gin.Context) {
	var body cancelBody
	_ = c.ShouldBindJSON(&body)
	job, err := h.uc.Cancel(c.GetString("userId"), c.Param("id"), body.Reason)
	h.respond(c, job, err)
}

// --- Message endpoints ---

type sendMessageBody struct {
	Content string `json:"content" binding:"required"`
}

func (h *JobHandler) SendMessage(c *gin.Context) {
	var body sendMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg, err := h.muc.Send(c.GetString("userId"), c.Param("id"), body.Content)
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *JobHandler) ListMessages(c *gin.Context) {
	msgs, err := h.muc.ListByRequest(c.GetString("userId"), c.Param("id"))
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

// --- helper ---

func (h *JobHandler) respond(c *gin.Context, job interface{}, err error) {
	if err != nil {
		c.JSON(httpStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

func httpStatus(err error) int {
	msg := err.Error()
	if len(msg) >= 9 && msg[:9] == "forbidden" {
		return http.StatusForbidden
	}
	if msg == "user not found" || msg == "request not found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
