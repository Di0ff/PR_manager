package e2e_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	apiURL     string
	adminToken string
	client     *http.Client
)

func TestMain(m *testing.M) {
	apiURL = os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8081"
	}

	adminToken = os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = "test-admin-secret-token"
	}

	client = &http.Client{Timeout: 10 * time.Second}

	if !waitForAPI(apiURL, 30*time.Second) {
		fmt.Println("API not ready, exiting")
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func waitForAPI(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("API is ready")
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func TestBusinessLogic_PRCreation_AutoAssignsReviewers(t *testing.T) {
	teamResp := createTeam(t, map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Dave", "is_active": true},
		},
	})
	require.Equal(t, "backend", teamResp["team"].(map[string]interface{})["team_name"])

	prResp := createPR(t, map[string]interface{}{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	})

	pr := prResp["pr"].(map[string]interface{})

	assert.Equal(t, "pr-1001", pr["pull_request_id"], "ID должен совпадать")
	assert.Equal(t, "OPEN", pr["status"], "Новый PR должен быть OPEN")
	assert.Equal(t, "u1", pr["author_id"], "Автор должен совпадать")

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok, "assigned_reviewers должен быть массивом")
	assert.LessOrEqual(t, len(reviewers), 2, "Не более 2 ревьюверов")
	assert.Greater(t, len(reviewers), 0, "Минимум 1 ревьювер должен быть назначен")

	for _, r := range reviewers {
		assert.NotEqual(t, "u1", r.(string), "Автор не может быть ревьювером")
	}

	assert.NotNil(t, pr["createdAt"], "Должно быть поле createdAt в camelCase")
}

func TestBusinessLogic_PRMerge_StateTransition(t *testing.T) {
	createTeam(t, map[string]interface{}{
		"team_name": "frontend",
		"members": []map[string]interface{}{
			{"user_id": "f1", "username": "Eve", "is_active": true},
			{"user_id": "f2", "username": "Frank", "is_active": true},
			{"user_id": "f3", "username": "Grace", "is_active": true},
		},
	})

	prResp := createPR(t, map[string]interface{}{
		"pull_request_id":   "pr-2001",
		"pull_request_name": "Fix bug",
		"author_id":         "f1",
	})

	pr := prResp["pr"].(map[string]interface{})
	assert.Equal(t, "OPEN", pr["status"])

	mergeResp := mergePR(t, "pr-2001")
	mergedPR := mergeResp["pr"].(map[string]interface{})

	assert.Equal(t, "MERGED", mergedPR["status"], "Статус должен измениться на MERGED")
	assert.NotNil(t, mergedPR["mergedAt"], "Должна быть дата мержа")

	reviewers := pr["assigned_reviewers"].([]interface{})
	if len(reviewers) > 0 {
		firstReviewer := reviewers[0].(string)

		resp, body := makeRequest(t, "POST", "/pullRequest/reassign", map[string]interface{}{
			"pull_request_id": "pr-2001",
			"old_user_id":     firstReviewer,
		}, nil)

		assert.Equal(t, http.StatusConflict, resp.StatusCode, "Должна быть ошибка 409")

		var errorResp map[string]interface{}
		json.Unmarshal([]byte(body), &errorResp)
		errorObj := errorResp["error"].(map[string]interface{})
		assert.Equal(t, "PR_MERGED", errorObj["code"], "Должен быть код PR_MERGED")
	}
}

func TestBusinessLogic_InactiveUser_NotAssignedAsReviewer(t *testing.T) {
	createTeam(t, map[string]interface{}{
		"team_name": "devops",
		"members": []map[string]interface{}{
			{"user_id": "d1", "username": "Henry", "is_active": true},
			{"user_id": "d2", "username": "Iris", "is_active": false},
			{"user_id": "d3", "username": "Jack", "is_active": true},
		},
	})

	prResp := createPR(t, map[string]interface{}{
		"pull_request_id":   "pr-3001",
		"pull_request_name": "Deploy script",
		"author_id":         "d1",
	})

	pr := prResp["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})

	for _, r := range reviewers {
		assert.NotEqual(t, "d2", r.(string), "Неактивный пользователь d2 не должен быть назначен")
	}

	if len(reviewers) > 0 {
		assert.Equal(t, "d3", reviewers[0].(string), "Должен быть назначен только d3")
	}
}

func TestBusinessLogic_ReassignReviewer_SelectsValidCandidate(t *testing.T) {
	createTeam(t, map[string]interface{}{
		"team_name": "mobile",
		"members": []map[string]interface{}{
			{"user_id": "m1", "username": "Kate", "is_active": true},
			{"user_id": "m2", "username": "Leo", "is_active": true},
			{"user_id": "m3", "username": "Mia", "is_active": true},
			{"user_id": "m4", "username": "Noah", "is_active": true},
		},
	})

	prResp := createPR(t, map[string]interface{}{
		"pull_request_id":   "pr-4001",
		"pull_request_name": "iOS update",
		"author_id":         "m1",
	})

	pr := prResp["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})
	require.Greater(t, len(reviewers), 0, "Должен быть хотя бы один ревьювер")

	oldReviewer := reviewers[0].(string)

	reassignResp := reassignReviewer(t, "pr-4001", oldReviewer)

	newReviewer := reassignResp["replaced_by"].(string)
	assert.NotEqual(t, oldReviewer, newReviewer, "Новый ревьювер должен отличаться")
	assert.NotEqual(t, "m1", newReviewer, "Автор не может стать ревьювером")

	updatedPR := reassignResp["pr"].(map[string]interface{})
	updatedReviewers := updatedPR["assigned_reviewers"].([]interface{})

	found := false
	for _, r := range updatedReviewers {
		if r.(string) == newReviewer {
			found = true
			break
		}
	}
	assert.True(t, found, "Новый ревьювер должен быть в списке assigned_reviewers")
}

func TestBusinessLogic_SetIsActive_RequiresAuth(t *testing.T) {
	createTeam(t, map[string]interface{}{
		"team_name": "security",
		"members": []map[string]interface{}{
			{"user_id": "s1", "username": "Olivia", "is_active": true},
		},
	})

	resp, body := makeRequest(t, "POST", "/users/setIsActive", map[string]interface{}{
		"user_id":   "s1",
		"is_active": false,
	}, nil)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Без токена должна быть ошибка 401")
	assert.Contains(t, body, "UNAUTHORIZED")

	headers := map[string]string{"Authorization": "Bearer wrong-token"}
	resp, body = makeRequest(t, "POST", "/users/setIsActive", map[string]interface{}{
		"user_id":   "s1",
		"is_active": false,
	}, headers)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "С неверным токеном должна быть ошибка 401")

	headers = map[string]string{"Authorization": "Bearer " + adminToken}
	resp, body = makeRequest(t, "POST", "/users/setIsActive", map[string]interface{}{
		"user_id":   "s1",
		"is_active": false,
	}, headers)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "С правильным токеном должно быть 200")

	var userResp map[string]interface{}
	json.Unmarshal([]byte(body), &userResp)
	user := userResp["user"].(map[string]interface{})
	assert.Equal(t, false, user["is_active"], "Пользователь должен быть деактивирован")
}

func createTeam(t *testing.T, data map[string]interface{}) map[string]interface{} {
	resp, body := makeRequest(t, "POST", "/team/add", data, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create team: %s", body)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(body), &result)
	require.NoError(t, err)
	return result
}

func createPR(t *testing.T, data map[string]interface{}) map[string]interface{} {
	resp, body := makeRequest(t, "POST", "/pullRequest/create", data, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create PR: %s", body)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(body), &result)
	require.NoError(t, err)
	return result
}

func mergePR(t *testing.T, prID string) map[string]interface{} {
	resp, body := makeRequest(t, "POST", "/pullRequest/merge", map[string]interface{}{
		"pull_request_id": prID,
	}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to merge PR: %s", body)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(body), &result)
	require.NoError(t, err)
	return result
}

func reassignReviewer(t *testing.T, prID, oldUserID string) map[string]interface{} {
	resp, body := makeRequest(t, "POST", "/pullRequest/reassign", map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     oldUserID,
	}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to reassign reviewer: %s", body)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(body), &result)
	require.NoError(t, err)
	return result
}

func makeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) (*http.Response, string) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, apiURL+path, reqBody)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(bodyBytes)
}
