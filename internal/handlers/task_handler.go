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

type TaskHandler struct {
	db       *gorm.DB
	validate *validator.Validate
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{
		db:       db,
		validate: validator.New(),
	}
}

// CreateTask creates a new task in a project
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid project ID",
			Code:    fiber.StatusBadRequest,
		})
	}

	var req models.TaskCreateRequest
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

	// Verify project exists and user owns it
	var project models.Project
	if err := h.db.Where("id = ? AND owner_id = ?", projectUUID, currentUserID).
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
			Message: "Failed to verify project",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Create task
	task := models.Task{
		Title:     req.Title,
		ProjectID: projectUUID,
		Status:    models.TaskStatusTodo,
		Priority:  models.TaskPriorityMedium,
	}

	if req.Description != "" {
		task.Description = &req.Description
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := h.db.Create(&task).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create task",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Load the task with relationships
	if err := h.db.Preload("Project").Preload("Assignee").First(&task, task.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to load task details",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		Message: "Task created successfully",
		Data:    task.ToResponse(),
	})
}

// GetProjectTasks retrieves tasks for a specific project
func (h *TaskHandler) GetProjectTasks(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	projectUUID, err := uuid.Parse(projectID)
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

	// Verify project exists and user owns it
	var project models.Project
	if err := h.db.Where("id = ? AND owner_id = ?", projectUUID, currentUserID).
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
			Message: "Failed to verify project",
			Code:    fiber.StatusInternalServerError,
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

	var tasks []models.Task
	var total int64

	// Count total tasks for the project
	if err := h.db.Model(&models.Task{}).Where("project_id = ?", projectUUID).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to count tasks",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Get tasks with pagination
	if err := h.db.Preload("Project").Preload("Assignee").
		Where("project_id = ?", projectUUID).
		Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch tasks",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Convert to response format
	taskResponses := make([]models.TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = task.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(models.ListResponse{
		Data: taskResponses,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetTask retrieves a task by ID
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	id := c.Params("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid task ID",
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

	var task models.Task
	if err := h.db.Preload("Project").Preload("Assignee").
		Joins("JOIN projects ON tasks.project_id = projects.id").
		Where("tasks.id = ? AND projects.owner_id = ?", taskID, currentUserID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "Task not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch task",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Task retrieved successfully",
		Data:    task.ToResponse(),
	})
}

// UpdateTask updates a task
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	id := c.Params("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid task ID",
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

	var req models.TaskUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	// Find task and verify ownership
	var task models.Task
	if err := h.db.Joins("JOIN projects ON tasks.project_id = projects.id").
		Where("tasks.id = ? AND projects.owner_id = ?", taskID, currentUserID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "Task not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch task",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Update fields
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := h.db.Save(&task).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update task",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Load the task with relationships
	if err := h.db.Preload("Project").Preload("Assignee").First(&task, task.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to load task details",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Task updated successfully",
		Data:    task.ToResponse(),
	})
}

// UpdateTaskStatus updates only the status of a task
func (h *TaskHandler) UpdateTaskStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid task ID",
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

	var req models.TaskStatusUpdateRequest
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

	// Find task and verify ownership
	var task models.Task
	if err := h.db.Joins("JOIN projects ON tasks.project_id = projects.id").
		Where("tasks.id = ? AND projects.owner_id = ?", taskID, currentUserID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "Not Found",
				Message: "Task not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch task",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Update status
	task.Status = req.Status

	if err := h.db.Save(&task).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update task status",
			Code:    fiber.StatusInternalServerError,
		})
	}

	// Load the task with relationships
	if err := h.db.Preload("Project").Preload("Assignee").First(&task, task.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to load task details",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Task status updated successfully",
		Data:    task.ToResponse(),
	})
}

// DeleteTask deletes a task
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid task ID",
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

	// Delete task with ownership verification
	result := h.db.Joins("JOIN projects ON tasks.project_id = projects.id").
		Where("tasks.id = ? AND projects.owner_id = ?", taskID, currentUserID).
		Delete(&models.Task{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete task",
			Code:    fiber.StatusInternalServerError,
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Not Found",
			Message: "Task not found",
			Code:    fiber.StatusNotFound,
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Task deleted successfully",
	})
}
