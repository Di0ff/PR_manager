package handlers

import (
	"mPR/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type API struct {
	services *service.Manager
	logger   *zap.Logger
}

func New(logger *zap.Logger, service *service.Manager) *API {
	return &API{
		services: service,
		logger:   logger,
	}
}

func (api *API) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
