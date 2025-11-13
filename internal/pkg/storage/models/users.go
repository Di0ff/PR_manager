package models

import (
	"time"

	"github.com/google/uuid"
)

type Users struct {
	ID        uuid.UUID `gorm:"column:user_id;primaryKey"`
	Username  string    `gorm:"column:username"`
	TeamName  *string   `gorm:"column:team_name"`
	IsActive  bool      `gorm:"column:is_active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	Team      *Teams    `gorm:"foreignKey:TeamName;references:Name"`
}
