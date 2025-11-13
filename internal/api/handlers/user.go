package handlers

import (
	"errors"
	"mPR/internal/api/dto"
	"mPR/internal/api/responses"
	"mPR/internal/custom"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (api *API) SetIsActive(c *gin.Context) {
	var input dto.SetIsActive
	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("неправильный json для SetIsActive", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	UUID, err := uuid.Parse(input.UserID)
	if err != nil {
		api.logger.Warn("неправильный user_id uuid", zap.String("user_id", input.UserID))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid user_id: must be UUID"))
		return
	}

	user, err := api.services.Users.SetActive(c, UUID, input.IsActive)
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

	uid, err := uuid.Parse(userID)
	if err != nil {
		api.logger.Warn("неправильный UUID для user_id", zap.String("user_id", userID))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid user_id"))
		return
	}

	prs, err := api.services.Users.GetUserReviews(c, uid)
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
