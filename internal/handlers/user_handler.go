package handlers

import (
	"math"
	"strconv"

	"taskflow-api/internal/middleware"
	"taskflow-api/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db       *gorm.DB
	validate *validator.Validate
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db:       db,
		validate: validator.New(),
	}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.UserCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Validation Error",
			Message: err.Error(),
			Code:    fiber.StatusBadRequest,
		})
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
			Error:   "Conflict",
			Message: "User with this email already exists",
			Code:    fiber.StatusConflict,
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to hash password",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if req.AvatarURL != "" {
		user.AvatarURL = &req.AvatarURL
	}

	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		Message: "User created successfully",
		Data:    user.ToResponse(),
	})
}

// GetUsers retrieves users with pagination
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var users []models.User
	var total int64

	// Count total users
	if err := h.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to count users",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Get users with pagination
	if err := h.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch users",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(models.ListResponse{
		Data: userResponses,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user.ToResponse(),
	})
}

// UpdateUser updates a user by ID
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Check if user is updating their own profile or has admin rights
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

	if currentUserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
			Error:   "Forbidden",
			Message: "You can only update your own profile",
			Code:    fiber.StatusForbidden,
		})
	}

	var req models.UserUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Find user
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "User updated successfully",
		Data:    user.ToResponse(),
	})
}

// DeleteUser soft deletes a user by ID
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Check if user is deleting their own profile or has admin rights
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

	if currentUserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
			Error:   "Forbidden",
			Message: "You can only delete your own profile",
			Code:    fiber.StatusForbidden,
		})
	}

	// Soft delete user
	if err := h.db.Delete(&models.User{}, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "User deleted successfully",
	})
}
