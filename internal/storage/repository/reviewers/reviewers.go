package reviewers

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

func (d *Database) Add(ctx context.Context, list []models.Reviewers) error {
	if len(list) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Create(&list).Error
}

func (d *Database) GetByPR(ctx context.Context, prID string) ([]models.Reviewers, error) {
	var reviewers []models.Reviewers
	err := d.db.WithContext(ctx).
		Where("pr_id = ?", prID).
		Find(&reviewers).Error

	return reviewers, err
}

func (d *Database) Delete(ctx context.Context, prID string, reviewerID string) error {
	return d.db.WithContext(ctx).
		Where("pr_id = ? AND reviewer_id = ?", prID, reviewerID).
		Delete(&models.Reviewers{}).Error
}

func (d *Database) AddOne(ctx context.Context, prID string, reviewerID string) error {
	return d.db.WithContext(ctx).
		Create(&models.Reviewers{
			PRID:       prID,
			ReviewerID: reviewerID,
		}).Error
}

func (d *Database) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]string, error) {
	var ids []string
	err := d.db.WithContext(ctx).
		Model(&models.Reviewers{}).
		Where("reviewer_id = ?", reviewerID).
		Pluck("pr_id", &ids).Error

	return ids, err
}
