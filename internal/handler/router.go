package handler

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/middleware"
)

func NewRouter(ph *ProfessionalHandler, oh *OnboardingHandler, mh *MeHandler, rh *RequestHandler, nh *NotificationHandler, ah *AdminHandler, jh *JobHandler, db *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://laboris-web.vercel.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Responde a todos los preflight OPTIONS
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.GET("/ping", Ping)

	// Rutas públicas
	pub := r.Group("/api/v1")
	{
		pub.GET("/professionals", ph.GetAll)
		pub.GET("/professionals/:id", ph.GetByID)
	}

	// Rutas protegidas — requieren JWT válido de Clerk
	priv := r.Group("/api/v1")
	priv.Use(middleware.ClerkAuth())
	{
		priv.POST("/onboarding", oh.Complete)
		priv.GET("/me/professional", mh.GetMyProfessional)
		priv.PUT("/me/professional", mh.UpdateMyProfessional)
		priv.POST("/requests", rh.Create)
		priv.GET("/me/requests/received", rh.ListReceived)
		priv.GET("/me/requests/sent", rh.ListSent)
		priv.PATCH("/requests/:id", rh.UpdateStatus)

		if nh != nil {
			priv.GET("/me/notifications", nh.List)
			priv.GET("/me/notifications/unread-count", nh.UnreadCount)
			priv.POST("/me/notifications/read-all", nh.MarkAllRead)
		}

		if jh != nil {
			priv.GET("/me/jobs", jh.ListMyJobs)
			priv.GET("/jobs/:id", jh.GetJob)
			priv.PATCH("/jobs/:id/schedule-visit", jh.ScheduleVisit)
			priv.PATCH("/jobs/:id/confirm-visit", jh.ConfirmVisit)
			priv.PATCH("/jobs/:id/decline-visit", jh.DeclineVisit)
			priv.PATCH("/jobs/:id/visit-quote", jh.SubmitVisitQuote)
			priv.PATCH("/jobs/:id/skip-visit", jh.SkipVisit)
			priv.PATCH("/jobs/:id/pay-visit", jh.PayVisit)
			priv.PATCH("/jobs/:id/complete-visit", jh.CompleteVisit)
			priv.PATCH("/jobs/:id/work-quote", jh.SubmitWorkQuote)
			priv.PATCH("/jobs/:id/approve-work", jh.ApproveWorkQuote)
			priv.PATCH("/jobs/:id/start-work", jh.StartWork)
			priv.PATCH("/jobs/:id/deliver-work", jh.DeliverWork)
			priv.PATCH("/jobs/:id/request-rework", jh.RequestRework)
			priv.PATCH("/jobs/:id/rework-quote", jh.SubmitReworkQuote)
			priv.PATCH("/jobs/:id/approve-rework-quote", jh.ApproveReworkQuote)
			priv.PATCH("/jobs/:id/accept-rework", jh.AcceptRework)
			priv.PATCH("/jobs/:id/approve-delivery", jh.ApproveDelivery)
			priv.PATCH("/jobs/:id/cancel", jh.Cancel)

			priv.POST("/requests/:id/messages", jh.SendMessage)
			priv.GET("/requests/:id/messages", jh.ListMessages)
		}
	}

	if ah != nil && db != nil {
		adminGrp := r.Group("/api/v1/admin")
		adminGrp.Use(middleware.ClerkAuth(), middleware.AdminAuth(db))
		{
			adminGrp.GET("/users", ah.ListUsers)
			adminGrp.GET("/professionals", ah.ListProfessionals)
			adminGrp.PATCH("/professionals/:id/verify", ah.VerifyProfessional)
			adminGrp.PATCH("/professionals/:id/status", ah.SetProfessionalStatus)
			adminGrp.DELETE("/professionals/:id", ah.DeleteProfessional)
		}
	}

	return r
}
