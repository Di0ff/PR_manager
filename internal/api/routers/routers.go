package routers

import (
	"mPR/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func Init(api *handlers.API) *gin.Engine {
	router := gin.Default()

	router.GET("/health", api.Health)

	team := router.Group("/team")
	{
		team.POST("/add", api.AddTeam)
		team.GET("/get", api.GetTeam)
	}

	user := router.Group("/users")
	{
		user.POST("/setIsActive", api.SetIsActive)
		user.GET("/getReview", api.GetReview)
	}

	pr := router.Group("/pullRequest")
	{
		pr.POST("/create", api.Create)
		pr.POST("/merge", api.Merge)
		pr.POST("/reassign", api.Reassign)
	}

	return router
}
