package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mPR/internal/api/handlers"
	"mPR/internal/service"
	"mPR/internal/service/teams"
	"mPR/internal/storage/models"
	"mPR/mocks"
)

func TestAddTeam_Success(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	mockTeams.EXPECT().GetByName(mock.Anything, "backend").Return(nil, gorm.ErrRecordNotFound)
	mockTeams.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.Teams")).Return(nil)
	mockUsers.EXPECT().CreateOrUpdate(mock.Anything, "backend", mock.AnythingOfType("[]models.Users")).Return(nil)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/team/add", api.AddTeam)

	body := `{
		"team_name": "backend",
		"members": [
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	team, ok := response["team"].(map[string]interface{})
	assert.True(t, ok, "response should contain 'team' field")
	assert.Equal(t, "backend", team["team_name"])

	members, ok := team["members"].([]interface{})
	assert.True(t, ok, "team should contain 'members' array")
	assert.Len(t, members, 2)

	member1 := members[0].(map[string]interface{})
	assert.Equal(t, "u1", member1["user_id"])
	assert.Equal(t, "Alice", member1["username"])
	assert.Equal(t, true, member1["is_active"])
}

func TestAddTeam_TeamExists(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	existingTeam := &models.Teams{Name: "backend"}
	mockTeams.EXPECT().GetByName(mock.Anything, "backend").Return(existingTeam, nil)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/team/add", api.AddTeam)

	body := `{
		"team_name": "backend",
		"members": [
			{"user_id": "u1", "username": "Alice", "is_active": true}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "TEAM_EXISTS")
}

func TestAddTeam_InvalidJSON(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/team/add", api.AddTeam)

	body := `{"team_name": "backend", invalid json`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid JSON")
}

func TestAddTeam_EmptyUserID(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/team/add", api.AddTeam)

	body := `{
		"team_name": "backend",
		"members": [
			{"user_id": "", "username": "Alice", "is_active": true}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user_id is required")
}

func TestGetTeam_Success(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	teamName := "backend"
	team := &models.Teams{
		Name: teamName,
		Users: []models.Users{
			{ID: "u1", Username: "Alice", IsActive: true, TeamName: &teamName},
			{ID: "u2", Username: "Bob", IsActive: true, TeamName: &teamName},
		},
	}

	mockTeams.EXPECT().GetByName(mock.Anything, "backend").Return(team, nil)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.GET("/team/get", api.GetTeam)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "backend", response["team_name"])

	members, ok := response["members"].([]interface{})
	assert.True(t, ok, "response should contain 'members' array")
	assert.Len(t, members, 2)
}

func TestGetTeam_MissingTeamName(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.GET("/team/get", api.GetTeam)

	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "team_name is required")
}

func TestGetTeam_NotFound(t *testing.T) {
	mockTeams := mocks.NewMockTeams(t)
	mockUsers := mocks.NewMockUsers(t)

	mockTeams.EXPECT().GetByName(mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	teamService := teams.New(mockTeams, mockUsers)
	services := &service.Manager{Teams: teamService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.GET("/team/get", api.GetTeam)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}
