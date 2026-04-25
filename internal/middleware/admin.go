package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AdminAuth(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		clerkID := c.GetString("userId")
		if clerkID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var count int
		err := db.QueryRow(context.Background(), `
			SELECT COUNT(*) FROM user_roles ur
			JOIN users u ON u.id = ur.user_id
			WHERE u.clerk_id = $1 AND ur.role = 'admin'
		`, clerkID).Scan(&count)
		if err != nil || count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}
