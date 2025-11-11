package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MockAuthMiddleware обеспечивает фиктивную аутентификацию, читая заголовки запроса.
func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")
		orgID := c.GetHeader("X-Org-ID")

		if userID == "" || userRole == "" || orgID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Set("currentUserID", userID)
		c.Set("currentUserRole", userRole)
		c.Set("currentOrgID", orgID)

		c.Next()
	}
}
