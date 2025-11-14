package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mPR/internal/api/handlers"
	"mPR/internal/api/middleware"
	"mPR/internal/service"
	"mPR/internal/service/users"
	"mPR/internal/storage/models"
	"mPR/mocks"
)

func TestSetIsActive_Success(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	user := &models.Users{
		ID:       "u1",
		Username: "Alice",
		TeamName: stringPtr("backend"),
		IsActive: true,
	}

	mockUsers.EXPECT().GetByID(mock.Anything, "u1").Return(user, nil)
	mockUsers.EXPECT().UpdateIsActive(mock.Anything, "u1", false).Return(nil)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/users/setIsActive", middleware.AdminAuth("test-token"), api.SetIsActive)

	body := `{"user_id": "u1", "is_active": false}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	userResp, ok := response["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "u1", userResp["user_id"])
	assert.Equal(t, "Alice", userResp["username"])
	assert.Equal(t, "backend", userResp["team_name"])
}

func TestSetIsActive_Unauthorized(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/users/setIsActive", middleware.AdminAuth("test-token"), api.SetIsActive)

	body := `{"user_id": "u1", "is_active": false}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
}

func TestSetIsActive_InvalidToken(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/users/setIsActive", middleware.AdminAuth("test-token"), api.SetIsActive)

	body := `{"user_id": "u1", "is_active": false}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid admin token")
}

func TestSetIsActive_UserNotFound(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	mockUsers.EXPECT().GetByID(mock.Anything, "u999").Return(nil, gorm.ErrRecordNotFound)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/users/setIsActive", middleware.AdminAuth("test-token"), api.SetIsActive)

	body := `{"user_id": "u999", "is_active": false}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}

func TestGetReview_Success(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	user := &models.Users{
		ID:       "u2",
		Username: "Bob",
		IsActive: true,
	}

	now := time.Now()
	prIDs := []string{"pr-1001", "pr-1002"}
	prs := []models.PullRequests{
		{
			ID:        "pr-1001",
			Name:      "Feature A",
			AuthorID:  "u1",
			Status:    "OPEN",
			CreatedAt: now,
		},
		{
			ID:        "pr-1002",
			Name:      "Feature B",
			AuthorID:  "u3",
			Status:    "OPEN",
			CreatedAt: now.Add(-1 * time.Hour),
		},
	}

	mockUsers.EXPECT().GetByID(mock.Anything, "u2").Return(user, nil)
	mockReviewers.EXPECT().GetPRsByReviewer(mock.Anything, "u2").Return(prIDs, nil)
	mockPR.EXPECT().GetByID(mock.Anything, "pr-1001").Return(&prs[0], nil)
	mockPR.EXPECT().GetByID(mock.Anything, "pr-1002").Return(&prs[1], nil)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.GET("/users/getReview", api.GetReview)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "u2", response["user_id"])

	pullRequests, ok := response["pull_requests"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, pullRequests, 2)

	pr1 := pullRequests[0].(map[string]interface{})
	assert.Equal(t, "pr-1001", pr1["pull_request_id"])
	assert.Equal(t, "Feature A", pr1["pull_request_name"])
	assert.Equal(t, "u1", pr1["author_id"])
	assert.Equal(t, "OPEN", pr1["status"])

	_, hasCreatedAt := pr1["createdAt"]
	assert.True(t, hasCreatedAt, "should have createdAt field in camelCase")
}

func TestGetReview_MissingUserID(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.GET("/users/getReview", api.GetReview)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user_id is required")
}

func TestGetReview_UserNotFound(t *testing.T) {
	mockUsers := mocks.NewMockUsers(t)
	mockPR := mocks.NewMockPullRequests(t)
	mockReviewers := mocks.NewMockReviewers(t)

	mockUsers.EXPECT().GetByID(mock.Anything, "u999").Return(nil, gorm.ErrRecordNotFound)

	userService := users.New(mockUsers, mockPR, mockReviewers)
	services := &service.Manager{Users: userService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.GET("/users/getReview", api.GetReview)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}
