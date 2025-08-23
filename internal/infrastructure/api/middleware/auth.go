package middleware

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
)

const (
	FirebaseContextKey = "firebaseUser"
	FirebaseUIDKey     = "firebaseUID"
)

func FirebaseAuthentication(app *firebase.App, cfg *config.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		authClient, err := app.Auth(context.Background())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		idToken := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Firebase ID token is missing"})
			return
		}

		token, err := authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(FirebaseContextKey, token)
		c.Set(FirebaseUIDKey, token.UID)
		c.Next()
	}
}
