package users

import (
	"context"
	"errors"
	"fmt"
	models2 "mPR/internal/storage/models"
	"mPR/internal/storage/repository"

	"mPR/internal/custom"

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

func (s *Service) SetActive(ctx context.Context, userID string, active bool) (*models2.Users, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}

	if err := s.users.UpdateIsActive(ctx, userID, active); err != nil {
		return nil, fmt.Errorf("update user is_active status: %w", err)
	}

	user.IsActive = active
	return user, nil
}

func (s *Service) GetUserReviews(ctx context.Context, userID string) ([]models2.PullRequests, error) {
	_, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}

	prIDs, err := s.reviewers.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get pull requests by reviewer: %w", err)
	}

	prs := make([]models2.PullRequests, 0, len(prIDs))
	for _, id := range prIDs {
		pr, err := s.pullRequests.GetByID(ctx, id)
		if err == nil {
			prs = append(prs, *pr)
		}
	}

	return prs, nil
}
