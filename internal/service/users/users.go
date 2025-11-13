package users

import (
	"context"
	"errors"
	"mPR/internal/custom"
	"mPR/internal/pkg/storage/models"
	"mPR/internal/pkg/storage/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	users        repository.Users
	pullRequests repository.PullRequests
	reviewers    repository.Reviewers
}

func New(users repository.Users, pullRequests repository.PullRequests, reviewers repository.Reviewers) *Service {
	return &Service{
		users:        users,
		pullRequests: pullRequests,
		reviewers:    reviewers,
	}
}

func (s *Service) SetActive(ctx context.Context, userID uuid.UUID, active bool) (*models.Users, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, err
	}

	if err := s.users.UpdateIsActive(ctx, userID, active); err != nil {
		return nil, err
	}

	user.IsActive = active
	return user, nil
}

func (s *Service) GetUserReviews(ctx context.Context, userID uuid.UUID) ([]models.PullRequests, error) {
	_, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, err
	}

	prIDs, err := s.reviewers.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	prs := make([]models.PullRequests, 0, len(prIDs))
	for _, id := range prIDs {
		pr, err := s.pullRequests.GetByID(ctx, id)
		if err == nil {
			prs = append(prs, *pr)
		}
	}

	return prs, nil
}
