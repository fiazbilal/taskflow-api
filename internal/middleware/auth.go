package middleware

import (
	"strings"
	"time"

	"taskflow-api/internal/config"
	"taskflow-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func JWTMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Missing authorization header",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid authorization header format",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Validate token and extract claims
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// Check if token is expired
			if claims.ExpiresAt.Time.Before(time.Now()) {
				return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
					Error:   "Unauthorized",
					Message: "Token expired",
					Code:    fiber.StatusUnauthorized,
				})
			}

			// Set user info in context
			c.Locals("user_id", claims.UserID)
			c.Locals("user_email", claims.Email)

			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid token claims",
			Code:    fiber.StatusUnauthorized,
		})
	}
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(user *models.User, cfg *config.Config) (string, error) {
	// Parse JWT expiry duration
	duration, err := time.ParseDuration(cfg.JWT.Expiry)
	if err != nil {
		duration = 24 * time.Hour // Default to 24 hours
	}

	// Create claims
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GetUserIDFromContext extracts user ID from fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}
	return userID, nil
}

// GetUserEmailFromContext extracts user email from fiber context
func GetUserEmailFromContext(c *fiber.Ctx) (string, error) {
	email, ok := c.Locals("user_email").(string)
	if !ok {
		return "", fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}
	return email, nil
}
