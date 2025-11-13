package pullRequests

import (
	"context"
	"errors"
	"mPR/internal/custom"
	"mPR/internal/pkg/storage/models"
	"mPR/internal/pkg/storage/repository"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	pullRequests repository.PullRequests
	users        repository.Users
	reviewers    repository.Reviewers
}

func New(pullRequests repository.PullRequests, users repository.Users, reviewers repository.Reviewers) *Service {
	return &Service{
		pullRequests: pullRequests,
		users:        users,
		reviewers:    reviewers,
	}
}

func (s *Service) Create(ctx context.Context, pr *models.PullRequests) (*models.PullRequests, error) {
	exist, err := s.pullRequests.GetByID(ctx, pr.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if exist != nil {
		return nil, custom.ErrPRExists
	}

	author, err := s.users.GetByID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, err
	}

	selected, err := s.selectReviewers(ctx, author)
	if err != nil {
		return nil, err
	}

	if err := s.pullRequests.Create(ctx, pr); err != nil {
		return nil, err
	}

	for i := range selected {
		selected[i].PRID = pr.ID
	}

	if err := s.reviewers.Add(ctx, selected); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Service) selectReviewers(ctx context.Context, author *models.Users) ([]models.Reviewers, error) {
	if author.TeamName == nil {
		return nil, custom.ErrNotFound
	}

	users, err := s.users.GetActiveByTeam(ctx, *author.TeamName)
	if err != nil {
		return nil, err
	}

	filtered := make([]models.Users, 0, len(users))
	for _, u := range users {
		if u.ID != author.ID {
			filtered = append(filtered, u)
		}
	}

	if len(filtered) == 0 {
		return []models.Reviewers{}, nil
	}

	rand.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	limit := 2
	if len(filtered) < 2 {
		limit = len(filtered)
	}

	result := make([]models.Reviewers, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, models.Reviewers{
			ReviewerID: filtered[i].ID,
		})
	}

	return result, nil
}

func (s *Service) Merge(ctx context.Context, prID uuid.UUID) (*models.PullRequests, error) {
	pr, err := s.pullRequests.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, err
	}

	if pr.Status == custom.StatusMerged {
		return pr, nil
	}

	pr.Status = custom.StatusMerged
	now := time.Now()
	pr.MergedAt = &now

	if err := s.pullRequests.Update(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Service) Reassign(ctx context.Context, prID, oldID uuid.UUID) (*models.PullRequests, uuid.UUID, error) {
	pr, err := s.pullRequests.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, uuid.Nil, custom.ErrNotFound
		}
		return nil, uuid.Nil, err
	}

	if pr.Status == custom.StatusMerged {
		return nil, uuid.Nil, custom.ErrPRMerged
	}

	reviewers, err := s.reviewers.GetByPR(ctx, prID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	isAssigned := false
	for _, r := range reviewers {
		if r.ReviewerID == oldID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, uuid.Nil, custom.ErrNotAssigned
	}

	oldUser, err := s.users.GetByID(ctx, oldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, uuid.Nil, custom.ErrNotFound
		}
		return nil, uuid.Nil, err
	}

	if oldUser.TeamName == nil {
		return nil, uuid.Nil, custom.ErrNoCandidate
	}

	candidates, err := s.users.GetActiveByTeam(ctx, *oldUser.TeamName)
	if err != nil {
		return nil, uuid.Nil, err
	}

	used := map[uuid.UUID]struct{}{
		oldID:       {},
		pr.AuthorID: {},
	}

	for _, r := range reviewers {
		used[r.ReviewerID] = struct{}{}
	}

	free := make([]models.Users, 0, len(candidates))
	for _, c := range candidates {
		if _, banned := used[c.ID]; !banned {
			free = append(free, c)
		}
	}

	if len(free) == 0 {
		return nil, uuid.Nil, custom.ErrNoCandidate
	}

	rand.Shuffle(len(free), func(i, j int) {
		free[i], free[j] = free[j], free[i]
	})

	newReviewer := free[0].ID

	if err := s.reviewers.Delete(ctx, prID, oldID); err != nil {
		return nil, uuid.Nil, err
	}
	if err := s.reviewers.AddOne(ctx, prID, newReviewer); err != nil {
		return nil, uuid.Nil, err
	}

	updatedPR, err := s.pullRequests.GetByID(ctx, prID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return updatedPR, newReviewer, nil
}
