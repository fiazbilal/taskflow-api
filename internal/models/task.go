package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string
type TaskPriority string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

type Task struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string         `json:"title" gorm:"not null"`
	Description *string        `json:"description"`
	ProjectID   uuid.UUID      `json:"project_id" gorm:"type:uuid;not null;index"`
	AssigneeID  *uuid.UUID     `json:"assignee_id" gorm:"type:uuid;index"`
	Status      TaskStatus     `json:"status" gorm:"type:task_status;default:'todo'"`
	Priority    TaskPriority   `json:"priority" gorm:"type:task_priority;default:'medium'"`
	DueDate     *time.Time     `json:"due_date"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Project  Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Assignee *User   `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
}

type TaskCreateRequest struct {
	Title       string        `json:"title" validate:"required"`
	Description string        `json:"description,omitempty"`
	AssigneeID  *uuid.UUID    `json:"assignee_id,omitempty"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
}

type TaskUpdateRequest struct {
	Title       string        `json:"title,omitempty"`
	Description *string       `json:"description,omitempty"`
	AssigneeID  *uuid.UUID    `json:"assignee_id,omitempty"`
	Status      *TaskStatus   `json:"status,omitempty"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
}

type TaskStatusUpdateRequest struct {
	Status TaskStatus `json:"status" validate:"required"`
}

type TaskResponse struct {
	ID          uuid.UUID        `json:"id"`
	Title       string           `json:"title"`
	Description *string          `json:"description"`
	ProjectID   uuid.UUID        `json:"project_id"`
	AssigneeID  *uuid.UUID       `json:"assignee_id"`
	Status      TaskStatus       `json:"status"`
	Priority    TaskPriority     `json:"priority"`
	DueDate     *time.Time       `json:"due_date"`
	CompletedAt *time.Time       `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Project     *ProjectResponse `json:"project,omitempty"`
	Assignee    *UserResponse    `json:"assignee,omitempty"`
}

func (t *Task) ToResponse() TaskResponse {
	response := TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		ProjectID:   t.ProjectID,
		AssigneeID:  t.AssigneeID,
		Status:      t.Status,
		Priority:    t.Priority,
		DueDate:     t.DueDate,
		CompletedAt: t.CompletedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}

	if t.Project.ID != uuid.Nil {
		projectResponse := t.Project.ToResponse()
		response.Project = &projectResponse
	}

	if t.Assignee != nil && t.Assignee.ID != uuid.Nil {
		assigneeResponse := t.Assignee.ToResponse()
		response.Assignee = &assigneeResponse
	}

	return response
}

// BeforeUpdate hook to set completed_at when status changes to done
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	if t.Status == TaskStatusDone && t.CompletedAt == nil {
		now := time.Now()
		t.CompletedAt = &now
	} else if t.Status != TaskStatusDone {
		t.CompletedAt = nil
	}
	return nil
}
