package models

import (
	"time"

	"github.com/google/uuid"
)

type PullRequests struct {
	ID        uuid.UUID   `gorm:"column:pr_id;primaryKey"`
	Name      string      `gorm:"column:pr_name"`
	AuthorID  uuid.UUID   `gorm:"column:author_id"`
	Status    string      `gorm:"column:status"`
	CreatedAt time.Time   `gorm:"column:created_at"`
	MergedAt  *time.Time  `gorm:"column:merged_at"`
	Author    Users       `gorm:"foreignKey:AuthorID;references:ID"`
	Reviewers []Reviewers `gorm:"foreignKey:PRID;references:ID"`
}
