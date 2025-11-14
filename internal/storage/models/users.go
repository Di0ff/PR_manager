package models

import "time"

type Users struct {
	ID        string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username  string    `gorm:"column:username" json:"username"`
	TeamName  *string   `gorm:"column:team_name" json:"team_name,omitempty"`
	IsActive  bool      `gorm:"column:is_active" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	Team      *Teams    `gorm:"foreignKey:TeamName;references:Name" json:"-"`
}
