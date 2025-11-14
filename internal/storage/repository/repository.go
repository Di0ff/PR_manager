package repository

import (
	"context"

	"gorm.io/gorm"

	"mPR/internal/storage/models"
	"mPR/internal/storage/repository/pull_requests"
	"mPR/internal/storage/repository/reviewers"
	"mPR/internal/storage/repository/teams"
	"mPR/internal/storage/repository/users"
)

type All struct {
	Teams        Teams
	Users        Users
	PullRequests PullRequests
	Reviewers    Reviewers
}

func New(db *gorm.DB) *All {
	return &All{
		Teams:        teams.New(db),
		Users:        users.New(db),
		PullRequests: pull_requests.New(db),
		Reviewers:    reviewers.New(db),
	}
}

type Teams interface {
	Create(ctx context.Context, team *models.Teams) error
	GetByName(ctx context.Context, name string) (*models.Teams, error)
}

type Users interface {
	GetByID(ctx context.Context, id string) (*models.Users, error)
	GetActiveByTeam(ctx context.Context, team string) ([]models.Users, error)
	UpdateIsActive(ctx context.Context, id string, active bool) error
	CreateOrUpdate(ctx context.Context, teamName string, members []models.Users) error
}

type PullRequests interface {
	Create(ctx context.Context, pr *models.PullRequests) error
	GetByID(ctx context.Context, id string) (*models.PullRequests, error)
	Update(ctx context.Context, pr *models.PullRequests) error
	AddReviewers(ctx context.Context, reviewers []models.Reviewers) error
	GetReviewers(ctx context.Context, prID string) ([]models.Reviewers, error)
	ReplaceReviewer(ctx context.Context, prID string, oldID, newID string) error
	GetByReviewer(ctx context.Context, reviewerID string) ([]models.PullRequests, error)
}

type Reviewers interface {
	Add(ctx context.Context, list []models.Reviewers) error
	GetByPR(ctx context.Context, prID string) ([]models.Reviewers, error)
	Delete(ctx context.Context, prID string, reviewerID string) error
	AddOne(ctx context.Context, prID string, reviewerID string) error
	GetPRsByReviewer(ctx context.Context, reviewerID string) ([]string, error)
}
