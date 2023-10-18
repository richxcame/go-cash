package routes

import (
	"gocash/controllers"
	"gocash/middlewares"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())

	r.POST("/cashes", controllers.CashController)
	r.POST("/ranges", controllers.RangesController)
	r.GET("/ranges", middlewares.Auth(), controllers.RangesWithAuth)
	// /cashes
	// Filters: amount, detail, note, client, contact as array
	// Pagination: offset, limit with defaults respectively 0, 50
	r.GET("/cashes", middlewares.Auth(), controllers.CashesWihtAuth)
	r.POST("/login", controllers.Login)
	r.POST("/token", controllers.RefreshToken)
	r.GET("/cashes/:uuid", middlewares.Auth(), controllers.CashesByUUID)

	return r
}
