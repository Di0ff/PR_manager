package teams_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"mPR/internal/custom"
	"mPR/internal/service/teams"
	"mPR/internal/storage/models"
	"mPR/mocks"
)

func TestAdd_Success(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	service := teams.New(mockTeams, mockUsers)

	ctx := context.Background()
	teamName := "team1"
	team := &models.Teams{Name: teamName}
	members := []models.Users{
		{Username: "user1", IsActive: true},
		{Username: "user2", IsActive: true},
	}

	mockTeams.On("GetByName", ctx, teamName).Return(nil, gorm.ErrRecordNotFound)
	mockTeams.On("Create", ctx, team).Return(nil)
	mockUsers.On("CreateOrUpdate", ctx, teamName, members).Return(nil)

	err := service.Add(ctx, team, members)

	assert.NoError(t, err)
}

func TestAdd_TeamExists(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	service := teams.New(mockTeams, mockUsers)

	ctx := context.Background()
	teamName := "team1"
	team := &models.Teams{Name: teamName}
	members := []models.Users{
		{Username: "user1", IsActive: true},
	}

	existingTeam := &models.Teams{Name: teamName}
	mockTeams.On("GetByName", ctx, teamName).Return(existingTeam, nil)

	err := service.Add(ctx, team, members)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrTeamExists))
}

func TestAdd_CreateError(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	service := teams.New(mockTeams, mockUsers)

	ctx := context.Background()
	teamName := "team1"
	team := &models.Teams{Name: teamName}
	members := []models.Users{
		{Username: "user1", IsActive: true},
	}

	mockTeams.On("GetByName", ctx, teamName).Return(nil, gorm.ErrRecordNotFound)
	mockTeams.On("Create", ctx, team).Return(errors.New("db error"))

	err := service.Add(ctx, team, members)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestGet_Success(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	service := teams.New(mockTeams, mockUsers)

	ctx := context.Background()
	teamName := "team1"
	expectedTeam := &models.Teams{Name: teamName}

	mockTeams.On("GetByName", ctx, teamName).Return(expectedTeam, nil)

	result, err := service.Get(ctx, teamName)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, teamName, result.Name)
}

func TestGet_NotFound(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	service := teams.New(mockTeams, mockUsers)

	ctx := context.Background()
	teamName := "nonexistent"

	mockTeams.On("GetByName", ctx, teamName).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.Get(ctx, teamName)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, custom.ErrNotFound))
	assert.Nil(t, result)
}

func TestGet_DBError(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	service := teams.New(mockTeams, mockUsers)

	ctx := context.Background()
	teamName := "team1"

	mockTeams.On("GetByName", ctx, teamName).Return(nil, errors.New("connection error"))

	result, err := service.Get(ctx, teamName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection error")
	assert.Nil(t, result)
}
