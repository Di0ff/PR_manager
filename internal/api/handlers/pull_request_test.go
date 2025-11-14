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
	"mPR/internal/custom"
	"mPR/internal/service"
	"mPR/internal/service/pull_requests"
	"mPR/internal/storage/models"
	"mPR/mocks"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreatePR_Success(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	author := &models.Users{
		ID:       "u1",
		Username: "Alice",
		TeamName: stringPtr("backend"),
		IsActive: true,
	}

	teamMembers := []models.Users{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: true},
		{ID: "u3", Username: "Charlie", IsActive: true},
	}

	mockPR.EXPECT().GetByID(mock.Anything, "pr-1001").Return(nil, gorm.ErrRecordNotFound)
	mockUsers.EXPECT().GetByID(mock.Anything, "u1").Return(author, nil)
	mockUsers.EXPECT().GetActiveByTeam(mock.Anything, "backend").Return(teamMembers, nil)
	mockPR.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.PullRequests")).Return(nil)
	mockReviewers.EXPECT().Add(mock.Anything, mock.AnythingOfType("[]models.Reviewers")).Return(nil)

	prService := pull_requests.New(mockPR, mockUsers, mockReviewers, 2)
	services := &service.Manager{PullRequests: prService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/pullRequest/create", api.Create)

	body := `{"pull_request_id": "pr-1001", "pull_request_name": "Add feature", "author_id": "u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	assert.True(t, ok, "response should contain 'pr' field")

	assert.Equal(t, "pr-1001", pr["pull_request_id"])
	assert.Equal(t, "Add feature", pr["pull_request_name"])
	assert.Equal(t, "u1", pr["author_id"])
	assert.Equal(t, "OPEN", pr["status"])

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	assert.True(t, ok, "assigned_reviewers should be array")
	assert.LessOrEqual(t, len(reviewers), 2, "should have max 2 reviewers")

	_, hasCreatedAt := pr["createdAt"]
	assert.True(t, hasCreatedAt, "should have createdAt field")
}

func TestCreatePR_PRExists(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	existingPR := &models.PullRequests{ID: "pr-1001"}
	mockPR.EXPECT().GetByID(mock.Anything, "pr-1001").Return(existingPR, nil)

	prService := pull_requests.New(mockPR, mockUsers, mockReviewers, 2)
	services := &service.Manager{PullRequests: prService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/pullRequest/create", api.Create)

	body := `{"pull_request_id": "pr-1001", "pull_request_name": "Add feature", "author_id": "u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "PR_EXISTS")
}

func TestMergePR_Success(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	now := time.Now()
	pr := &models.PullRequests{
		ID:        "pr-1001",
		Name:      "Add feature",
		AuthorID:  "u1",
		Status:    custom.StatusOpen,
		CreatedAt: now.Add(-1 * time.Hour),
	}

	mockPR.EXPECT().GetByID(mock.Anything, "pr-1001").Return(pr, nil)
	mockPR.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p *models.PullRequests) bool {
		return p.Status == custom.StatusMerged && p.MergedAt != nil
	})).Return(nil)

	prService := pull_requests.New(mockPR, mockUsers, mockReviewers, 2)
	services := &service.Manager{PullRequests: prService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/pullRequest/merge", api.Merge)

	body := `{"pull_request_id": "pr-1001"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	prResponse, ok := response["pr"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "MERGED", prResponse["status"])

	_, hasMergedAt := prResponse["mergedAt"]
	assert.True(t, hasMergedAt, "should have mergedAt field")
}

func TestReassignPR_Success(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	pr := &models.PullRequests{
		ID:       "pr-1001",
		Name:     "Add feature",
		AuthorID: "u1",
		Status:   custom.StatusOpen,
	}

	reviewers := []models.Reviewers{
		{PRID: "pr-1001", ReviewerID: "u2"},
		{PRID: "pr-1001", ReviewerID: "u3"},
	}

	oldUser := &models.Users{
		ID:       "u2",
		Username: "Bob",
		TeamName: stringPtr("backend"),
		IsActive: true,
	}

	candidates := []models.Users{
		{ID: "u2", Username: "Bob", IsActive: true},
		{ID: "u3", Username: "Charlie", IsActive: true},
		{ID: "u4", Username: "Dave", IsActive: true},
	}

	mockPR.EXPECT().GetByID(mock.Anything, "pr-1001").Return(pr, nil).Times(2)
	mockReviewers.EXPECT().GetByPR(mock.Anything, "pr-1001").Return(reviewers, nil)
	mockUsers.EXPECT().GetByID(mock.Anything, "u2").Return(oldUser, nil)
	mockUsers.EXPECT().GetActiveByTeam(mock.Anything, "backend").Return(candidates, nil)
	mockReviewers.EXPECT().Delete(mock.Anything, "pr-1001", "u2").Return(nil)
	mockReviewers.EXPECT().AddOne(mock.Anything, "pr-1001", "u4").Return(nil)

	prService := pull_requests.New(mockPR, mockUsers, mockReviewers, 2)
	services := &service.Manager{PullRequests: prService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/pullRequest/reassign", api.Reassign)

	body := `{"pull_request_id": "pr-1001", "old_user_id": "u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	replacedBy, ok := response["replaced_by"].(string)
	assert.True(t, ok, "response should contain replaced_by field")
	assert.NotEmpty(t, replacedBy)
	assert.Equal(t, "u4", replacedBy)
}

func TestReassignPR_PRMerged(t *testing.T) {
	mockPR := mocks.NewMockPullRequests(t)
	mockUsers := mocks.NewMockUsers(t)
	mockReviewers := mocks.NewMockReviewers(t)

	pr := &models.PullRequests{
		ID:       "pr-1001",
		AuthorID: "u1",
		Status:   custom.StatusMerged,
	}

	mockPR.EXPECT().GetByID(mock.Anything, "pr-1001").Return(pr, nil)

	prService := pull_requests.New(mockPR, mockUsers, mockReviewers, 2)
	services := &service.Manager{PullRequests: prService}
	api := handlers.New(zap.NewNop(), services)

	router := gin.New()
	router.POST("/pullRequest/reassign", api.Reassign)

	body := `{"pull_request_id": "pr-1001", "old_user_id": "u2"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "PR_MERGED")
}

func stringPtr(s string) *string {
	return &s
}
