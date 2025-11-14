package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing Authorization header",
				},
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid Authorization header format",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token != adminToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid admin token",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
