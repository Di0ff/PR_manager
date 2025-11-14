package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"mPR/internal/api/dto"
	"mPR/internal/api/responses"
	"mPR/internal/custom"
	"mPR/internal/storage/models"
)

func (api *API) Create(c *gin.Context) {
	var input dto.CreatePR

	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("неправильный json для Create", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	if input.PullRequestID == "" {
		api.logger.Warn("пустой pull_request_id")
		c.JSON(http.StatusBadRequest, responses.Error("", "pull_request_id is required"))
		return
	}

	if input.AuthorID == "" {
		api.logger.Warn("пустой author_id")
		c.JSON(http.StatusBadRequest, responses.Error("", "author_id is required"))
		return
	}

	pr := &models.PullRequests{
		ID:       input.PullRequestID,
		Name:     input.PullRequestName,
		AuthorID: input.AuthorID,
		Status:   custom.StatusOpen,
	}

	create, err := api.services.PullRequests.Create(c, pr)
	if err != nil {
		if errors.Is(err, custom.ErrPRExists) {
			c.JSON(http.StatusConflict,
				responses.Error("PR_EXISTS", "PR id already exists"),
			)
			return
		}

		if errors.Is(err, custom.ErrNotFound) {
			c.JSON(http.StatusNotFound,
				responses.Error("NOT_FOUND", "author or team not found"),
			)
			return
		}

		api.logger.Error("ошибка создания PR", zap.Error(err))
		c.JSON(http.StatusInternalServerError,
			responses.Error("", "internal server error"),
		)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": create})
}

func (api *API) Merge(c *gin.Context) {
	var input dto.Merge

	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("неправильный json для Merge", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	if input.PRID == "" {
		api.logger.Warn("пустой pull_request_id")
		c.JSON(http.StatusBadRequest, responses.Error("", "pull_request_id is required"))
		return
	}

	pr, err := api.services.PullRequests.Merge(c, input.PRID)
	if err != nil {
		if errors.Is(err, custom.ErrNotFound) {
			c.JSON(http.StatusNotFound,
				responses.Error("NOT_FOUND", "resource not found"),
			)
			return
		}

		api.logger.Error("ошибка merge PR", zap.Error(err))
		c.JSON(http.StatusInternalServerError,
			responses.Error("", "internal server error"),
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

func (api *API) Reassign(c *gin.Context) {
	var input dto.ReassignRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("неправильный json для Reassign", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	if input.PullRequestID == "" {
		api.logger.Warn("пустой pull_request_id")
		c.JSON(http.StatusBadRequest, responses.Error("", "pull_request_id is required"))
		return
	}

	if input.OldUserID == "" {
		api.logger.Warn("пустой old_user_id")
		c.JSON(http.StatusBadRequest, responses.Error("", "old_user_id is required"))
		return
	}

	pr, newReviewerID, err := api.services.PullRequests.Reassign(c, input.PullRequestID, input.OldUserID)
	if err != nil {
		if errors.Is(err, custom.ErrNotFound) {
			c.JSON(http.StatusNotFound,
				responses.Error("NOT_FOUND", "PR or user not found"),
			)
			return
		}

		if errors.Is(err, custom.ErrPRMerged) {
			c.JSON(http.StatusConflict,
				responses.Error("PR_MERGED", "cannot reassign on merged PR"),
			)
			return
		}

		if errors.Is(err, custom.ErrNotAssigned) {
			c.JSON(http.StatusConflict,
				responses.Error("NOT_ASSIGNED", "reviewer is not assigned to this PR"),
			)
			return
		}

		if errors.Is(err, custom.ErrNoCandidate) {
			c.JSON(http.StatusConflict,
				responses.Error("NO_CANDIDATE", "no active replacement candidate in team"),
			)
			return
		}

		api.logger.Error("ошибка reassign reviewer", zap.Error(err))
		c.JSON(http.StatusInternalServerError,
			responses.Error("", "internal server error"),
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
