package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	// Context key under which we store the user’s UUID
	UserIDKey = "userID"

	// Fallback secret — in prod you should ALWAYS set JWT_SECRET instead
	JwtSecret = "verylongrandomstringyouwritehere-and-never-commit-an-obvious-password"
)

// JWT_SECRET env var name
const EnvJWTSecret = "JWT_SECRET"

// JWTAuthMiddleware validates the token and pulls the "sub" claim into context.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Load secret from env or fallback
		secret := []byte(os.Getenv(EnvJWTSecret))
		if len(secret) == 0 {
			secret = []byte(JwtSecret)
		}

		// Parse & validate token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract claims and pull out the subject
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sub claim not found in token"})
			return
		}

		// Store the user's UUID in context
		c.Set(UserIDKey, sub)
		c.Next()
	}
}

// GetUserID retrieves the userID (the JWT "sub" claim) from context.
// Returns the ID string and a boolean if it was present.
func GetUserID(c *gin.Context) (string, bool) {
	v, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	userID, ok := v.(string)
	return userID, ok
}
