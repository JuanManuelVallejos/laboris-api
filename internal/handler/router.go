package handler

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laboris/laboris-api/internal/middleware"
)

func NewRouter(ph *ProfessionalHandler, oh *OnboardingHandler, mh *MeHandler, rh *RequestHandler, nh *NotificationHandler, ah *AdminHandler, db *pgxpool.Pool) *gin.Engine {
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
