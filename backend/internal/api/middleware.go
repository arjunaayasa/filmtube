package api

import (
	"net/http"
	"strings"

	"github.com/arjunaayasa/filmtube/internal/auth"
	"github.com/arjunaayasa/filmtube/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserKey contextKey = "user"
	UserIDKey contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserRoleKey), claims.Role)
		c.Set(string(UserKey), claims)

		c.Next()
	}
}

// RequireCreator middleware ensures user has creator or admin role
func RequireCreator() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(UserRoleKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userRole := role.(models.UserRole)
		if !auth.IsCreator(userRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "creator access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware ensures user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(UserRoleKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userRole := role.(models.UserRole)
		if !auth.IsAdmin(userRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

// GetUserRole retrieves user role from context
func GetUserRole(c *gin.Context) (models.UserRole, bool) {
	role, exists := c.Get(string(UserRoleKey))
	if !exists {
		return "", false
	}
	return role.(models.UserRole), true
}
