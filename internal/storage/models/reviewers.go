package models

type Reviewers struct {
	PRID       string `gorm:"column:pr_id;primaryKey" json:"pr_id"`
	ReviewerID string `gorm:"column:reviewer_id;primaryKey" json:"reviewer_id"`
}
