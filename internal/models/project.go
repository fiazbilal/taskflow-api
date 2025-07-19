package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusArchived  ProjectStatus = "archived"
	ProjectStatusCompleted ProjectStatus = "completed"
)

type Project struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description *string        `json:"description"`
	Color       string         `json:"color" gorm:"default:'#6366f1'"`
	OwnerID     uuid.UUID      `json:"owner_id" gorm:"type:uuid;not null;index"`
	Status      ProjectStatus  `json:"status" gorm:"type:project_status;default:'active'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Owner User   `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ProjectID"`
}

type ProjectCreateRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

type ProjectUpdateRequest struct {
	Name        string         `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Color       string         `json:"color,omitempty"`
	Status      *ProjectStatus `json:"status,omitempty"`
}

type ProjectResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Color       string        `json:"color"`
	OwnerID     uuid.UUID     `json:"owner_id"`
	Status      ProjectStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Owner       *UserResponse `json:"owner,omitempty"`
	TasksCount  int           `json:"tasks_count,omitempty"`
}

type ProjectWithTasksResponse struct {
	ProjectResponse
	Tasks []TaskResponse `json:"tasks"`
}

func (p *Project) ToResponse() ProjectResponse {
	response := ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Color:       p.Color,
		OwnerID:     p.OwnerID,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.Owner.ID != uuid.Nil {
		ownerResponse := p.Owner.ToResponse()
		response.Owner = &ownerResponse
	}

	response.TasksCount = len(p.Tasks)

	return response
}

func (p *Project) ToResponseWithTasks() ProjectWithTasksResponse {
	response := ProjectWithTasksResponse{
		ProjectResponse: p.ToResponse(),
		Tasks:           make([]TaskResponse, len(p.Tasks)),
	}

	for i, task := range p.Tasks {
		response.Tasks[i] = task.ToResponse()
	}

	return response
}
