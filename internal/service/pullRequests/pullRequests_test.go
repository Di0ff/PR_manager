package pullRequests

import (
	"context"
	"errors"
	"mPR/internal/custom"
	"mPR/internal/pkg/storage/models"
	"mPR/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCreate_Success(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	authorID := uuid.New()
	teamName := "team1"

	pr := &models.PullRequests{
		ID:       prID,
		Name:     "Test PR",
		AuthorID: authorID,
		Status:   custom.StatusOpen,
	}

	author := &models.Users{
		ID:       authorID,
		Username: "author",
		TeamName: &teamName,
		IsActive: true,
	}

	reviewer1ID := uuid.New()
	reviewer2ID := uuid.New()
	activeUsers := []models.Users{
		{ID: authorID, Username: "author", IsActive: true, TeamName: &teamName},
		{ID: reviewer1ID, Username: "reviewer1", IsActive: true, TeamName: &teamName},
		{ID: reviewer2ID, Username: "reviewer2", IsActive: true, TeamName: &teamName},
	}

	mockPR.On("GetByID", ctx, prID).Return(nil, gorm.ErrRecordNotFound)
	mockUsers.On("GetByID", ctx, authorID).Return(author, nil)
	mockUsers.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)
	mockPR.On("Create", ctx, pr).Return(nil)
	mockReviewers.On("Add", ctx, mock.AnythingOfType("[]models.Reviewers")).Return(nil)

	result, err := service.Create(ctx, pr)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, prID, result.ID)
}

func TestCreate_PRExists(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	authorID := uuid.New()

	pr := &models.PullRequests{
		ID:       prID,
		Name:     "Test PR",
		AuthorID: authorID,
		Status:   custom.StatusOpen,
	}

	existingPR := &models.PullRequests{ID: prID}
	mockPR.On("GetByID", ctx, prID).Return(existingPR, nil)

	result, err := service.Create(ctx, pr)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrPRExists))
	assert.Nil(t, result)
}

func TestCreate_AuthorNotFound(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	authorID := uuid.New()

	pr := &models.PullRequests{
		ID:       prID,
		Name:     "Test PR",
		AuthorID: authorID,
		Status:   custom.StatusOpen,
	}

	mockPR.On("GetByID", ctx, prID).Return(nil, gorm.ErrRecordNotFound)
	mockUsers.On("GetByID", ctx, authorID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.Create(ctx, pr)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNotFound))
	assert.Nil(t, result)
}

func TestMerge_Success(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()

	pr := &models.PullRequests{
		ID:     prID,
		Name:   "Test PR",
		Status: custom.StatusOpen,
	}

	mockPR.On("GetByID", ctx, prID).Return(pr, nil)
	mockPR.On("Update", ctx, mock.MatchedBy(func(p *models.PullRequests) bool {
		return p.ID == prID && p.Status == custom.StatusMerged && p.MergedAt != nil
	})).Return(nil)

	result, err := service.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, custom.StatusMerged, result.Status)
	assert.NotNil(t, result.MergedAt)
}

func TestMerge_AlreadyMerged(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	mergedAt := time.Now()

	pr := &models.PullRequests{
		ID:       prID,
		Name:     "Test PR",
		Status:   custom.StatusMerged,
		MergedAt: &mergedAt,
	}

	mockPR.On("GetByID", ctx, prID).Return(pr, nil)

	result, err := service.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, custom.StatusMerged, result.Status)
}

func TestMerge_PRNotFound(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()

	mockPR.On("GetByID", ctx, prID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.Merge(ctx, prID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNotFound))
	assert.Nil(t, result)
}

func TestReassign_Success(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	oldReviewerID := uuid.New()
	newReviewerID := uuid.New()
	authorID := uuid.New()
	teamName := "team1"

	pr := &models.PullRequests{
		ID:       prID,
		Name:     "Test PR",
		AuthorID: authorID,
		Status:   custom.StatusOpen,
	}

	oldReviewer := &models.Users{
		ID:       oldReviewerID,
		Username: "old_reviewer",
		TeamName: &teamName,
		IsActive: true,
	}

	reviewers := []models.Reviewers{
		{ReviewerID: oldReviewerID, PRID: prID},
	}

	activeUsers := []models.Users{
		{ID: authorID, Username: "author", IsActive: true, TeamName: &teamName},
		{ID: oldReviewerID, Username: "old_reviewer", IsActive: true, TeamName: &teamName},
		{ID: newReviewerID, Username: "new_reviewer", IsActive: true, TeamName: &teamName},
	}

	mockPR.On("GetByID", ctx, prID).Return(pr, nil).Once()
	mockReviewers.On("GetByPR", ctx, prID).Return(reviewers, nil)
	mockUsers.On("GetByID", ctx, oldReviewerID).Return(oldReviewer, nil)
	mockUsers.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)
	mockReviewers.On("Delete", ctx, prID, oldReviewerID).Return(nil)
	mockReviewers.On("AddOne", ctx, prID, mock.AnythingOfType("uuid.UUID")).Return(nil)
	mockPR.On("GetByID", ctx, prID).Return(pr, nil).Once()

	result, replacedBy, err := service.Reassign(ctx, prID, oldReviewerID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, replacedBy)
}

func TestReassign_PRMerged(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	oldReviewerID := uuid.New()

	pr := &models.PullRequests{
		ID:     prID,
		Name:   "Test PR",
		Status: custom.StatusMerged,
	}

	mockPR.On("GetByID", ctx, prID).Return(pr, nil)

	result, replacedBy, err := service.Reassign(ctx, prID, oldReviewerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrPRMerged))
	assert.Nil(t, result)
	assert.Equal(t, uuid.Nil, replacedBy)
}

func TestReassign_NotAssigned(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	oldReviewerID := uuid.New()
	otherReviewerID := uuid.New()

	pr := &models.PullRequests{
		ID:     prID,
		Name:   "Test PR",
		Status: custom.StatusOpen,
	}

	reviewers := []models.Reviewers{
		{ReviewerID: otherReviewerID, PRID: prID},
	}

	mockPR.On("GetByID", ctx, prID).Return(pr, nil)
	mockReviewers.On("GetByPR", ctx, prID).Return(reviewers, nil)

	result, replacedBy, err := service.Reassign(ctx, prID, oldReviewerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNotAssigned))
	assert.Nil(t, result)
	assert.Equal(t, uuid.Nil, replacedBy)
}

func TestReassign_NoCandidate(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	service := New(mockPR, mockUsers, mockReviewers)

	ctx := context.Background()
	prID := uuid.New()
	oldReviewerID := uuid.New()
	authorID := uuid.New()
	teamName := "team1"

	pr := &models.PullRequests{
		ID:       prID,
		Name:     "Test PR",
		AuthorID: authorID,
		Status:   custom.StatusOpen,
	}

	oldReviewer := &models.Users{
		ID:       oldReviewerID,
		Username: "old_reviewer",
		TeamName: &teamName,
		IsActive: true,
	}

	reviewers := []models.Reviewers{
		{ReviewerID: oldReviewerID, PRID: prID},
	}

	activeUsers := []models.Users{
		{ID: authorID, Username: "author", IsActive: true, TeamName: &teamName},
		{ID: oldReviewerID, Username: "old_reviewer", IsActive: true, TeamName: &teamName},
	}

	mockPR.On("GetByID", ctx, prID).Return(pr, nil)
	mockReviewers.On("GetByPR", ctx, prID).Return(reviewers, nil)
	mockUsers.On("GetByID", ctx, oldReviewerID).Return(oldReviewer, nil)
	mockUsers.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil)

	result, replacedBy, err := service.Reassign(ctx, prID, oldReviewerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNoCandidate))
	assert.Nil(t, result)
	assert.Equal(t, uuid.Nil, replacedBy)
}
