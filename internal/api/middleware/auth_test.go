package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"mPR/internal/api/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAdminAuth_Success(t *testing.T) {
	router := gin.New()
	router.POST("/test", middleware.AdminAuth("secret-token"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestAdminAuth_MissingHeader(t *testing.T) {
	router := gin.New()
	router.POST("/test", middleware.AdminAuth("secret-token"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
	assert.Contains(t, w.Body.String(), "missing Authorization header")
}

func TestAdminAuth_InvalidFormat(t *testing.T) {
	router := gin.New()
	router.POST("/test", middleware.AdminAuth("secret-token"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "secret-token"},
		{"Wrong prefix", "Basic secret-token"},
		{"Empty token", "Bearer "},
		{"Only Bearer", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
		})
	}
}

func TestAdminAuth_InvalidToken(t *testing.T) {
	router := gin.New()
	router.POST("/test", middleware.AdminAuth("secret-token"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
	assert.Contains(t, w.Body.String(), "invalid admin token")
}

func TestAdminAuth_EmptyToken(t *testing.T) {
	router := gin.New()
	router.POST("/test", middleware.AdminAuth(""), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
