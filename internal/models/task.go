package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid();column:id"`
	Title       string         `json:"title" gorm:"not null;column:title" validate:"required,min=1,max=255"`
	Description *string        `json:"description,omitempty" gorm:"column:description"`
	Completed   bool           `json:"completed" gorm:"default:false;column:completed"`
	DueDate     time.Time      `json:"dueDate" gorm:"type:date;not null;column:due_date" validate:"required"`
	Priority    string         `json:"priority" gorm:"default:'medium';column:priority" validate:"oneof=low medium high urgent"`
	Category    string         `json:"category" gorm:"default:'personal';column:category" validate:"oneof=personal office learning research"`
	Status      string         `json:"status" gorm:"default:'pending';column:status" validate:"oneof=pending in-progress review completed cancelled"`
	ProjectID   *uuid.UUID     `json:"projectId,omitempty" gorm:"type:uuid;column:project_id"`
	Tags        pq.StringArray `json:"tags" gorm:"type:text[];column:tags"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"column:updated_at"`
}

// TableName specifies the table name for GORM
func (Task) TableName() string {
	return "tasks"
}

// BeforeCreate hook to set default values
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description *string    `json:"description"`
	DueDate     time.Time  `json:"dueDate" validate:"required"`
	Priority    string     `json:"priority" validate:"oneof=low medium high urgent"`
	Category    string     `json:"category" validate:"oneof=personal office learning research"`
	Status      string     `json:"status" validate:"oneof=pending in-progress review completed cancelled"`
	ProjectID   *uuid.UUID `json:"projectId"`
	Tags        []string   `json:"tags"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description"`
	Completed   *bool      `json:"completed"`
	DueDate     *time.Time `json:"dueDate"`
	Priority    *string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	Category    *string    `json:"category" validate:"omitempty,oneof=personal office learning research"`
	Status      *string    `json:"status" validate:"omitempty,oneof=pending in-progress review completed cancelled"`
	ProjectID   *uuid.UUID `json:"projectId"`
	Tags        *[]string  `json:"tags"`
}

// TaskFilters represents query parameters for filtering tasks
type TaskFilters struct {
	Category   string `form:"category"`
	Priority   string `form:"priority"`
	Status     string `form:"status"`
	Completed  *bool  `form:"completed"`
	Search     string `form:"search"`
	DateFilter string `form:"dateFilter"` // today, tomorrow, this-week, overdue
	ProjectID  string `form:"projectId"`
	Tags       string `form:"tags"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
	SortBy     string `form:"sortBy"`    // dueDate, priority, created, title, category
	SortOrder  string `form:"sortOrder"` // asc, desc
}
