package routes

import (
	"gocash/middlewares"
	"gocash/src/auth"
	"gocash/src/booking"
	"gocash/src/cash"
	"gocash/src/ranges"
	"gocash/src/transaction"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())
	r.POST("/cashes", cash.CreateCashController)
	r.POST("/ranges", ranges.CreateRangesController)
	r.POST("/trxns", middlewares.Auth(), transaction.Transaction)
	r.GET("/ranges", middlewares.Auth(), ranges.GetRanges)
	r.GET("/cashes", middlewares.Auth(), cash.GetCashes)
	r.GET("/check/:booking-number", middlewares.Auth(), booking.CheckBookingNumber)
	r.POST("/login", auth.Login)
	r.POST("/token", auth.RefreshToken)
	r.GET("/cashes/:booking-number", middlewares.Auth(), cash.CashesByBooking)
	r.GET("/dates-cashes", middlewares.Auth(), cash.GetDatesCashes)

	return r
}
