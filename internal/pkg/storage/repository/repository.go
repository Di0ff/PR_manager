package repository

import (
	"context"
	"mPR/internal/pkg/storage/models"
	"mPR/internal/pkg/storage/repository/pullRequests"
	"mPR/internal/pkg/storage/repository/reviewers"
	"mPR/internal/pkg/storage/repository/teams"
	"mPR/internal/pkg/storage/repository/users"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
		PullRequests: pullRequests.New(db),
		Reviewers:    reviewers.New(db),
	}
}

type Teams interface {
	Create(ctx context.Context, team *models.Teams) error
	GetByName(ctx context.Context, name string) (*models.Teams, error)
}

type Users interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Users, error)
	GetActiveByTeam(ctx context.Context, team string) ([]models.Users, error)
	UpdateIsActive(ctx context.Context, id uuid.UUID, active bool) error
	CreateOrUpdate(ctx context.Context, teamName string, members []models.Users) error
}

type PullRequests interface {
	Create(ctx context.Context, pr *models.PullRequests) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PullRequests, error)
	Update(ctx context.Context, pr *models.PullRequests) error
	AddReviewers(ctx context.Context, reviewers []models.Reviewers) error
	GetReviewers(ctx context.Context, prID uuid.UUID) ([]models.Reviewers, error)
	ReplaceReviewer(ctx context.Context, prID uuid.UUID, oldID, newID uuid.UUID) error
	GetByReviewer(ctx context.Context, reviewerID uuid.UUID) ([]models.PullRequests, error)
}

type Reviewers interface {
	Add(ctx context.Context, list []models.Reviewers) error
	GetByPR(ctx context.Context, prID uuid.UUID) ([]models.Reviewers, error)
	Delete(ctx context.Context, prID uuid.UUID, reviewerID uuid.UUID) error
	AddOne(ctx context.Context, prID uuid.UUID, reviewerID uuid.UUID) error
	GetPRsByReviewer(ctx context.Context, reviewerID uuid.UUID) ([]uuid.UUID, error)
}
