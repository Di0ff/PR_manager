package pullRequests

import (
	"context"
	"mPR/internal/pkg/storage/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Database {
	return &Database{
		db: db,
	}
}

func (d *Database) Create(ctx context.Context, pr *models.PullRequests) error {
	return d.db.WithContext(ctx).Create(pr).Error
}

func (d *Database) GetByID(ctx context.Context, id uuid.UUID) (*models.PullRequests, error) {
	var pr models.PullRequests
	err := d.db.WithContext(ctx).
		Preload("Author").
		Preload("Reviewers").
		First(&pr, "pr_id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (d *Database) Update(ctx context.Context, pr *models.PullRequests) error {
	return d.db.WithContext(ctx).Save(pr).Error
}

func (d *Database) AddReviewers(ctx context.Context, reviewers []models.Reviewers) error {
	return d.db.WithContext(ctx).Create(&reviewers).Error
}

func (d *Database) GetReviewers(ctx context.Context, prID uuid.UUID) ([]models.Reviewers, error) {
	var list []models.Reviewers
	err := d.db.WithContext(ctx).
		Where("pr_id = ?", prID).
		Find(&list).Error

	return list, err
}

func (d *Database) ReplaceReviewer(ctx context.Context, prID uuid.UUID, oldID, newID uuid.UUID) error {
	err := d.db.WithContext(ctx).
		Where("pr_id = ? AND reviewer_id = ?", prID, oldID).
		Delete(&models.Reviewers{}).Error

	if err != nil {
		return err
	}

	return d.db.WithContext(ctx).
		Create(&models.Reviewers{
			PRID:       prID,
			ReviewerID: newID,
		}).Error
}

func (d *Database) GetByReviewer(ctx context.Context, reviewerID uuid.UUID) ([]models.PullRequests, error) {
	var prs []models.PullRequests
	err := d.db.WithContext(ctx).
		Joins("JOIN pr_reviewers r ON r.pr_id = pull_requests.pr_id").
		Where("r.reviewer_id = ?", reviewerID).
		Find(&prs).Error

	return prs, err
}
