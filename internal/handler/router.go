package handler

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/laboris/laboris-api/internal/middleware"
)

func NewRouter(ph *ProfessionalHandler, oh *OnboardingHandler, mh *MeHandler) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://laboris-web.vercel.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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
	}

	return r
}
