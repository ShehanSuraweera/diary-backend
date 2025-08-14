package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Resource struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid();column:id"`
	Title         string         `json:"title" gorm:"not null;column:title"`
	URL           string         `json:"url" gorm:"column:url"`
	Description   string         `json:"description" gorm:"column:description"`
	Technology    string         `json:"technology" gorm:"not null;column:technology"`
	Type          string         `json:"type" gorm:"not null;column:type"`
	Status        string         `json:"status" gorm:"column:status"`
	Priority      string         `json:"priority" gorm:"column:priority"`
	Rating        *int           `json:"rating" gorm:"column:rating"`
	EstimatedTime *int           `json:"estimated_time" gorm:"column:estimated_time"`
	Progress      *int           `json:"progress" gorm:"column:progress"`
	Notes         string         `json:"notes" gorm:"column:notes"`
	Tags          pq.StringArray `json:"tags" gorm:"type:text[];column:tags"`
	CompletedAt   *time.Time     `json:"completed_at" gorm:"column:completed_at"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at"`
}

func (Resource) TableName() string {
	return "learning_resources"
}
