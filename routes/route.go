package routes

import (
	"gocash/controllers"
	"gocash/middlewares"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())

	r.POST("/cashes", controllers.CreateCashController)
	r.POST("/ranges", controllers.CreateRangesController)
	r.GET("/ranges", middlewares.Auth(), controllers.GetRangesWithAuth)
	r.GET("/cashes", middlewares.Auth(), controllers.GetCashesWihtAuth)
	r.POST("/login", controllers.Login)
	r.POST("/token", controllers.RefreshToken)
	r.GET("/cashes/:uuid", middlewares.Auth(), controllers.CashesByUUID)

	return r
}
