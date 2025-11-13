package teams

import (
	"context"
	"errors"
	"mPR/internal/custom"
	"mPR/internal/pkg/storage/models"
	"mPR/internal/pkg/storage/repository"

	"gorm.io/gorm"
)

type Service struct {
	teams repository.Teams
	users repository.Users
}

func New(teams repository.Teams, users repository.Users) *Service {
	return &Service{
		teams: teams,
		users: users,
	}
}

func (t *Service) Add(ctx context.Context, team *models.Teams, members []models.Users) error {
	exist, err := t.teams.GetByName(ctx, team.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if exist != nil {
		return custom.ErrTeamExists
	}

	if err := t.teams.Create(ctx, team); err != nil {
		return err
	}

	if err := t.users.CreateOrUpdate(ctx, team.Name, members); err != nil {
		return err
	}

	return nil
}

func (t *Service) Get(ctx context.Context, name string) (*models.Teams, error) {
	team, err := t.teams.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom.ErrNotFound
		}
		return nil, err
	}

	return team, nil
}
