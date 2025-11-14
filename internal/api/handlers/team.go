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

func (api *API) AddTeam(c *gin.Context) {
	var input dto.Team
	if err := c.ShouldBindJSON(&input); err != nil {
		api.logger.Warn("Wrong json for AddTeam", zap.Error(err))
		c.JSON(http.StatusBadRequest, responses.Error("", "invalid JSON"))
		return
	}

	users := make([]models.Users, 0, len(input.Members))
	for _, m := range input.Members {
		if m.UserID == "" {
			api.logger.Warn("Empty user_id у member")
			c.JSON(http.StatusBadRequest, responses.Error("", "user_id is required"))
			return
		}

		users = append(users, models.Users{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
			TeamName: &input.TeamName,
		})
	}

	team := models.Teams{Name: input.TeamName}

	err := api.services.Teams.Add(c, &team, users)
	if err != nil {
		if errors.Is(err, custom.ErrTeamExists) {
			c.JSON(http.StatusBadRequest,
				responses.Error("TEAM_EXISTS", "team_name already exists"),
			)
			return
		}

		api.logger.Error("Error add team", zap.Error(err))
		c.JSON(http.StatusInternalServerError, responses.Error("", "internal server error"))
		return
	}

	team.Users = users
	c.JSON(http.StatusCreated, gin.H{"team": team})
}

func (api *API) GetTeam(c *gin.Context) {
	name := c.Query("team_name")
	if name == "" {
		api.logger.Warn("Missing team_name")
		c.JSON(http.StatusBadRequest, responses.Error("", "team_name is required"))
		return
	}

	team, err := api.services.Teams.Get(c, name)
	if err != nil {
		if errors.Is(err, custom.ErrNotFound) {
			c.JSON(http.StatusNotFound,
				responses.Error("NOT_FOUND", "team not found"),
			)
			return
		}

		api.logger.Error("Error get team", zap.Error(err))
		c.JSON(http.StatusInternalServerError, responses.Error("", "internal server error"))
		return
	}

	c.JSON(http.StatusOK, team)
}
