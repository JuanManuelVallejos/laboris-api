package handler

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(ph *ProfessionalHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", Ping)

	// Rutas públicas
	pub := r.Group("/api/v1")
	{
		pub.GET("/professionals", ph.GetAll)
		pub.GET("/professionals/:id", ph.GetByID)
	}

	// Rutas protegidas se agregan en el próximo paso (POST /requests, etc.)
	// Ejemplo:
	// priv := r.Group("/api/v1")
	// priv.Use(middleware.ClerkAuth())
	// priv.POST("/requests", rh.Create)

	return r
}
