package users

import (
	"context"
	"mPR/internal/pkg/storage/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Database struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}

func (d *Database) GetByID(ctx context.Context, id uuid.UUID) (*models.Users, error) {
	var user models.Users
	if err := d.db.WithContext(ctx).
		First(&user, "user_id = ?", id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *Database) GetActiveByTeam(ctx context.Context, team string) ([]models.Users, error) {
	var users []models.Users
	if err := d.db.WithContext(ctx).
		Where("team_name = ? AND is_active = true", team).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (d *Database) UpdateIsActive(ctx context.Context, id uuid.UUID, active bool) error {
	if err := d.db.WithContext(ctx).
		Model(&models.Users{}).
		Where("user_id = ?", id).
		Update("is_active", active).
		Error; err != nil {

		return err
	}

	return nil
}

func (d *Database) CreateOrUpdate(ctx context.Context, teamName string, members []models.Users) error {
	for i := range members {
		members[i].TeamName = &teamName
	}

	if err := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"username", "team_name", "is_active"}),
		}).
		Create(&members).Error; err != nil {

		return err
	}

	return nil
}
