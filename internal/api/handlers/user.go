package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"mPR/internal/api/dto"
	"mPR/internal/api/responses"
	"mPR/internal/custom"
)

func (api *API) SetIsActive(c *gin.Context) {
	var input dto.SetIsActive
	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("неправильный json для SetIsActive", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	if input.UserID == "" {
		api.logger.Warn("пустой user_id")
		c.JSON(http.StatusBadRequest, responses.Error("", "user_id is required"))
		return
	}

	user, err := api.services.Users.SetActive(c, input.UserID, input.IsActive)
	if err != nil {
		if errors.Is(err, custom.ErrNotFound) {
			c.JSON(http.StatusNotFound, responses.Error("NOT_FOUND", "user not found"))
			return
		}

		api.logger.Error("failed to set is_active", zap.Error(err))
		c.JSON(http.StatusInternalServerError, responses.Error("", "internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (api *API) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		api.logger.Warn("отсутствует user_id для GetReview")
		c.JSON(http.StatusBadRequest, responses.Error("", "user_id is required"))
		return
	}

	prs, err := api.services.Users.GetUserReviews(c, userID)
	if err != nil {
		if errors.Is(err, custom.ErrNotFound) {
			c.JSON(http.StatusNotFound, responses.Error("NOT_FOUND", "user not found"))
			return
		}

		api.logger.Error("ошибка получения reviews для user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, responses.Error("", "internal server error"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
