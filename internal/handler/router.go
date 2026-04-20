package handler

import "github.com/gin-gonic/gin"

func NewRouter(ph *ProfessionalHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", Ping)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/professionals", ph.GetAll)
		v1.GET("/professionals/:id", ph.GetByID)
	}

	return r
}
