package middleware

import (
	"net/http"
	"subservice/utils"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from the auth middleware (should be called before this)
		_, exists := c.Get("user_id")
		if !exists {
			utils.UnauthorizedResponse(c, "Authentication required")
			c.Abort()
			return
		}

		// For now, we'll check if user_id corresponds to admin username
		// In a real app, you'd query the database to get user details
		// But since we're keeping it simple, we'll check the username from token
		username, exists := c.Get("username")
		if !exists || username != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin privileges required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
