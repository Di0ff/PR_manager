package handlers

import (
	"errors"
	"mPR/internal/api/dto"
	"mPR/internal/api/responses"
	"mPR/internal/custom"
	"mPR/internal/pkg/storage/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (api *API) Create(c *gin.Context) {
	var input dto.CreatePR

	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("неправильный json для Create", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	UUID, err := uuid.Parse(input.PullRequestID)
	if err != nil {
		api.logger.Warn("неправильный UUID для pr_id",
			zap.String("id", input.PullRequestID),
		)
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid pull_request_id"))
		return
	}

	authorUUID, err := uuid.Parse(input.AuthorID)
	if err != nil {
		api.logger.Warn("неправильный UUID для author_id",
			zap.String("id", input.AuthorID),
		)
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid author_id"))
		return
	}

	pr := &models.PullRequests{
		ID:       UUID,
		Name:     input.PullRequestName,
		AuthorID: authorUUID,
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

	UUID, err := uuid.Parse(input.PRID)
	if err != nil {
		api.logger.Warn("неправильный UUID для pr_id",
			zap.String("id", input.PRID),
		)
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid pull_request_id"))
		return
	}

	pr, err := api.services.PullRequests.Merge(c, UUID)
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

	UUID, err := uuid.Parse(input.PullRequestID)
	if err != nil {
		api.logger.Warn("неправильный UUID для pr_id",
			zap.String("id", input.PullRequestID),
		)
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid pull_request_id"))
		return
	}

	oldUUID, err := uuid.Parse(input.OldUserID)
	if err != nil {
		api.logger.Warn("неправильный UUID для old_user_id",
			zap.String("id", input.OldUserID),
		)
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid old_user_id"))
		return
	}

	pr, newReviewerID, err := api.services.PullRequests.Reassign(c, UUID, oldUUID)
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
		"replaced_by": newReviewerID.String(),
	})
}
