package ranges

import (
	"gocash/models"
	"gocash/pkg/logger"
	"gocash/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateRangesController(c *gin.Context) {

	ctx := c.Request.Context()

	// Get request body
	var body models.RangeBody
	if err := c.BindJSON(&body); err != nil {
		logger.Errorf("request body wrong %v", err)
		c.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Request body invalid",
		})
		return
	}

	// Insert request to database
	_uuid, msg, err := CreateRangesService(ctx, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": msg,
		})
		return
	}

	// Send success result
	c.JSON(201, gin.H{
		"message": msg,
		"uuid":    _uuid,
	})
}

func GetRanges(c *gin.Context) {
	ctx := c.Request.Context()
	offset, limit := utils.Paginate(c)
	resultRanges, totalRanges, errMsh, msg := GetRangesService(ctx, offset, limit)
	if errMsh != "" {
		logger.Errorf(msg+" %v", errMsh)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errMsh,
			"message": msg,
		})
	}
	c.JSON(200, gin.H{
		"ranges": resultRanges,
		"total":  totalRanges,
	})
}
