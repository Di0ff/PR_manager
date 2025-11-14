package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type PullRequests struct {
	ID        string      `gorm:"column:pr_id;primaryKey" json:"pull_request_id"`
	Name      string      `gorm:"column:pr_name" json:"pull_request_name"`
	AuthorID  string      `gorm:"column:author_id" json:"author_id"`
	Status    string      `gorm:"column:status" json:"status"`
	CreatedAt time.Time   `gorm:"column:created_at" json:"createdAt"`
	MergedAt  *time.Time  `gorm:"column:merged_at" json:"mergedAt,omitempty"`
	Author    Users       `gorm:"foreignKey:AuthorID;references:ID" json:"-"`
	Reviewers []Reviewers `gorm:"foreignKey:PRID;references:ID" json:"-"`
}

func (pr *PullRequests) MarshalJSON() ([]byte, error) {
	type Alias PullRequests

	reviewerIDs := make([]string, 0, len(pr.Reviewers))
	for _, r := range pr.Reviewers {
		reviewerIDs = append(reviewerIDs, r.ReviewerID)
	}

	data, err := json.Marshal(&struct {
		*Alias
		AssignedReviewers []string `json:"assigned_reviewers"`
	}{
		Alias:             (*Alias)(pr),
		AssignedReviewers: reviewerIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal pull request JSON: %w", err)
	}
	return data, nil
}
