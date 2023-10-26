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
	r.POST("/trnxs", middlewares.Auth(), controllers.Transaction)
	r.GET("/ranges", middlewares.Auth(), controllers.GetRanges)
	r.GET("/cashes", middlewares.Auth(), controllers.GetCashes)
	r.GET("/check/:booking-number", middlewares.Auth(), controllers.CheckBookingNumber)
	r.POST("/login", controllers.Login)
	r.POST("/token", controllers.RefreshToken)
	r.GET("/cashes/:booking-number", middlewares.Auth(), controllers.CashesByBooking)

	return r
}
