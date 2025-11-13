package models

import "github.com/google/uuid"

type Reviewers struct {
	PRID       uuid.UUID `gorm:"column:pr_id;primaryKey"`
	ReviewerID uuid.UUID `gorm:"column:reviewer_id;primaryKey"`
}
