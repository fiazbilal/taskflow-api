package handlers

import (
	"math"
	"strconv"

	"taskflow-api/internal/middleware"
	"taskflow-api/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	db       *gorm.DB
	validate *validator.Validate
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{
		db:       db,
		validate: validator.New(),
	}
}

// CreateProject creates a new project
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req models.ProjectCreateRequest
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

	// Get current user
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

	// Create project
	project := models.Project{
		Name:    req.Name,
		OwnerID: currentUserID,
		Status:  models.ProjectStatusActive,
	}

	if req.Description != "" {
		project.Description = &req.Description
	}

	if req.Color != "" {
		project.Color = req.Color
	}

	if err := h.db.Create(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create project",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Load the project with owner
	if err := h.db.Preload("Owner").First(&project, project.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to load project details",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		Message: "Project created successfully",
		Data:    project.ToResponse(),
	})
}

// GetProjects retrieves user's projects with pagination
func (h *ProjectHandler) GetProjects(c *fiber.Ctx) error {
	// Get current user
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

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

	var projects []models.Project
	var total int64

	// Count total projects for the user
	if err := h.db.Model(&models.Project{}).Where("owner_id = ?", currentUserID).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to count projects",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Get projects with pagination
	if err := h.db.Preload("Owner").Preload("Tasks").
		Where("owner_id = ?", currentUserID).
		Offset(offset).Limit(limit).Find(&projects).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch projects",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Convert to response format
	projectResponses := make([]models.ProjectResponse, len(projects))
	for i, project := range projects {
		projectResponses[i] = project.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(models.ListResponse{
		Data: projectResponses,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetProject retrieves a project with its tasks
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	id := c.Params("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid project ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Get current user
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

	var project models.Project
	if err := h.db.Preload("Owner").Preload("Tasks").Preload("Tasks.Assignee").
		Where("id = ? AND owner_id = ?", projectID, currentUserID).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "Project not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch project",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Project retrieved successfully",
		Data:    project.ToResponseWithTasks(),
	})
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	id := c.Params("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid project ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Get current user
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

	var req models.ProjectUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Find project
	var project models.Project
	if err := h.db.Where("id = ? AND owner_id = ?", projectID, currentUserID).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "Project not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch project",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Update fields
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}
	if req.Color != "" {
		project.Color = req.Color
	}
	if req.Status != nil {
		project.Status = *req.Status
	}

	if err := h.db.Save(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update project",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Load the project with owner
	if err := h.db.Preload("Owner").First(&project, project.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to load project details",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Project updated successfully",
		Data:    project.ToResponse(),
	})
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	id := c.Params("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid project ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Get current user
	currentUserID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    fiber.StatusUnauthorized,
		})
	}

	// Delete project (this will also delete associated tasks due to foreign key constraints)
	result := h.db.Where("id = ? AND owner_id = ?", projectID, currentUserID).
		Delete(&models.Project{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete project",
			Code:    fiber.StatusInternalServerError,
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Not Found",
			Message: "Project not found",
			Code:    fiber.StatusNotFound,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Project deleted successfully",
	})
}
