package middleware

import (
	"net/http"
	"go-com/tokens"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieves the token from the request header
		ClientToken := c.Request.Header.Get("token")

		// If no authorization provided, respond with an error message and 
		// Abort() to ensure no other handlers are called by this request and return from the function
		if ClientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No authorization header provided"})
			c.Abort()
			return
		}

		// If the token is invalid, respond with an error message abort the context and return
		claims, err := tokens.ValidateToken(ClientToken)
		if err!="" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return 
		}

		// Stores new key-value pairs exclusively for this context
		c.Set("email", claims.Email)
		c.Set("uid", claims.Uid)

		// Executes the pending handlers inside the chain inside the calling handler
		c.Next()

	}
}