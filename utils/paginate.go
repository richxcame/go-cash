package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func Paginate(ctx *gin.Context) (offset, limit int) {
	offset, limit = 0, 20
	// Prepare pagination details
	offsetQuery := ctx.DefaultQuery("offset", "0")
	limitQuery := ctx.DefaultQuery("limit", "20")

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		offset = 0
	}
	limit, err = strconv.Atoi(limitQuery)
	if err != nil {
		limit = 50
	}
	return offset, limit
}
