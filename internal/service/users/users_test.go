package users

import (
	"context"
	"errors"
	"mPR/internal/custom"
	"mPR/internal/pkg/storage/models"
	"mPR/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSetActive_Success(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()
	user := &models.Users{
		ID:       userID,
		Username: "testuser",
		IsActive: false,
	}

	mockUsers.On("GetByID", ctx, userID).Return(user, nil)
	mockUsers.On("UpdateIsActive", ctx, userID, true).Return(nil)

	result, err := service.SetActive(ctx, userID, true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsActive)
	assert.Equal(t, userID, result.ID)
}

func TestSetActive_UserNotFound(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()

	mockUsers.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.SetActive(ctx, userID, true)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNotFound))
	assert.Nil(t, result)
}

func TestSetActive_UpdateError(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()
	user := &models.Users{
		ID:       userID,
		Username: "testuser",
		IsActive: false,
	}

	mockUsers.On("GetByID", ctx, userID).Return(user, nil)
	mockUsers.On("UpdateIsActive", ctx, userID, true).Return(errors.New("update failed"))

	result, err := service.SetActive(ctx, userID, true)

	assert.Error(t, err)
	assert.Equal(t, "update failed", err.Error())
	assert.Nil(t, result)
}

func TestGetUserReviews_Success(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()
	user := &models.Users{
		ID:       userID,
		Username: "reviewer",
		IsActive: true,
	}

	prID1 := uuid.New()
	prID2 := uuid.New()
	prIDs := []uuid.UUID{prID1, prID2}

	pr1 := &models.PullRequests{
		ID:     prID1,
		Name:   "PR 1",
		Status: custom.StatusOpen,
	}
	pr2 := &models.PullRequests{
		ID:     prID2,
		Name:   "PR 2",
		Status: custom.StatusMerged,
	}

	mockUsers.On("GetByID", ctx, userID).Return(user, nil)
	mockReviewers.On("GetPRsByReviewer", ctx, userID).Return(prIDs, nil)
	mockPR.On("GetByID", ctx, prID1).Return(pr1, nil)
	mockPR.On("GetByID", ctx, prID2).Return(pr2, nil)

	result, err := service.GetUserReviews(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, prID1, result[0].ID)
	assert.Equal(t, prID2, result[1].ID)
}

func TestGetUserReviews_UserNotFound(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()

	mockUsers.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.GetUserReviews(ctx, userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNotFound))
	assert.Nil(t, result)
}

func TestGetUserReviews_NoPRs(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()
	user := &models.Users{
		ID:       userID,
		Username: "reviewer",
		IsActive: true,
	}

	mockUsers.On("GetByID", ctx, userID).Return(user, nil)
	mockReviewers.On("GetPRsByReviewer", ctx, userID).Return([]uuid.UUID{}, nil)

	result, err := service.GetUserReviews(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestGetUserReviews_SkipsMissingPRs(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockUsers, mockPR, mockReviewers)

	ctx := context.Background()
	userID := uuid.New()
	user := &models.Users{
		ID:       userID,
		Username: "reviewer",
		IsActive: true,
	}

	prID1 := uuid.New()
	prID2 := uuid.New()
	prIDs := []uuid.UUID{prID1, prID2}

	pr1 := &models.PullRequests{
		ID:     prID1,
		Name:   "PR 1",
		Status: custom.StatusOpen,
	}

	mockUsers.On("GetByID", ctx, userID).Return(user, nil)
	mockReviewers.On("GetPRsByReviewer", ctx, userID).Return(prIDs, nil)
	mockPR.On("GetByID", ctx, prID1).Return(pr1, nil)
	mockPR.On("GetByID", ctx, prID2).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.GetUserReviews(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, prID1, result[0].ID)
}
