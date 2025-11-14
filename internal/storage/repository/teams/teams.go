package teams

import (
	"context"

	"gorm.io/gorm"

	"mPR/internal/storage/models"
)

type Database struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}

func (d *Database) Create(ctx context.Context, team *models.Teams) error {
	return d.db.WithContext(ctx).Create(team).Error
}

func (d *Database) GetByName(ctx context.Context, name string) (*models.Teams, error) {
	var team models.Teams
	if err := d.db.WithContext(ctx).
		Preload("Users").
		First(&team, "team_name = ?", name).Error; err != nil {
		return nil, err
	}

	return &team, nil
}
