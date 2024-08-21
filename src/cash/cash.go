package cash

import (
	"context"
	"fmt"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"gocash/utils"
	"gocash/utils/arrs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateCashController(ctx *gin.Context) {
	// Get request body
	var body models.CashBody
	if err := ctx.BindJSON(&body); err != nil {
		logger.Errorf("request body wrong %v", err)
		ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Request body invalid",
		})
		return
	}

	// Find the client with the given key
	var client string
	err := db.DB.QueryRow(context.Background(), slClsSql, body.APIKey).Scan(&client)
	if err != nil {
		logger.Errorf("api key search error %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err,
			"message": "Client hasn't been found",
		})
		return
	}

	// Insert request to database
	_uuid := uuid.New().String()
	_, err = db.DB.Exec(context.Background(), insCshSql, _uuid, time.Now(), time.Now(), client, body.Contact, body.Amount, body.Detail, body.Note)
	if err != nil {
		logger.Errorf("database save error %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err,
			"message": "Couldn't save into cashes",
		})
		return
	}

	// Send success result
	ctx.JSON(201, gin.H{
		"message": "Successfully saved into database",
		"uuid":    _uuid,
	})
}

func GetCashes(ctx *gin.Context) {
	offset, limit := utils.Paginate(ctx)

	urlQueries := ctx.Request.URL.Query()
	index := 0
	var values []interface{}
	var queries []string
	for k, v := range urlQueries {
		if arrs.Contains([]string{"uuid", "client", "contact", "amount", "detail", "note"}, k) {
			if k == "contact" {
				for kcon, vcon := range v {
					if strings.Contains(vcon, " ") {
						str, _ := url.QueryUnescape(strings.Split(vcon, " ")[1])
						v[kcon] = str
					}
				}
			}
			str := ""
			for _, v := range v {
				str += v + "|"
			}
			str = strings.TrimSuffix(str, "|")
			str += ""
			values = append(values, str)
			index++

			queries = append(queries, fmt.Sprintf("%s ~* $", k)+strconv.Itoa(index))
		}
	}
	valuesWithPagination := append(values, offset, limit)

	sqlFilters := ""
	if len(queries) > 0 {
		sqlFilters += " WHERE "
		sqlFilters += strings.Join(queries, " AND ")
	}
	slSumCshsSql += sqlFilters
	slSumCshsSql += " GROUP BY c.contact, c.detail ORDER BY MAX(created_at) DESC "
	slSumCshsSql += fmt.Sprintf(" offset $%v limit $%v", index+1, index+2)
	rows, err := db.DB.Query(context.Background(), slSumCshsSql, valuesWithPagination...)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Couldn't search from cashes",
		})
		return
	}
	defer rows.Close()

	cashes := make([]models.CashBodyResponse, 0)
	for rows.Next() {
		var cash models.CashBodyResponse
		err := rows.Scan(&cash.Contact, &cash.Detail, &cash.Amount, &cash.CreatedAt)
		if err != nil {
			logger.Errorf("Scan error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Scan error",
			})
			return
		}
		cashes = append(cashes, cash)
	}

	totalCashes := 0
	query := slCntCshsSql + sqlFilters + " GROUP BY contact, detail) AS subquery;"

	err = db.DB.QueryRow(context.Background(), query, values...).Scan(&totalCashes)

	if err != nil {
		logger.Errorf("cash count error %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Couldn't count total number of transactions",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"cashes": cashes,
		"total":  totalCashes,
	})
}

func CashesByBooking(ctx *gin.Context) {
	// Get Booking Number from URL param
	bookingNumber, ok := ctx.Params.Get("booking-number")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Booking Number is required",
			"message": "Coulnd't find Booking Number",
		})
		return
	}

	// Find the cash with given Booking Number
	cashes := make([]models.CashBodyResponseBooking, 0)
	rows, err := db.DB.Query(context.Background(), slCshsSql, bookingNumber)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   err.Error(),
			"message": "Couldn't find the cash details",
		})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cash models.CashBodyResponseBooking
		err := rows.Scan(&cash.UUID, &cash.CreatedAt, &cash.Client, &cash.Contact, &cash.Amount, &cash.Detail, &cash.Note)
		if err != nil {
			logger.Errorf("Scan error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Scan error",
			})
			return
		}
		cashes = append(cashes, cash)
	}
	ctx.JSON(200, gin.H{
		"cashes": cashes,
	})
}

func GetDatesCashes(c *gin.Context) {
	ctx := c.Request.Context()
	date1, date2 := getMonth()
	date1 = c.DefaultQuery("date1", date1)
	date2 = c.DefaultQuery("date2", date2)

	data, msg, err := GetDatesCashesService(&ctx, date1, date2)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": msg,
		})
		return
	}

	c.JSON(http.StatusOK, data)

}

func getMonth() (string, string) {
	now := time.Now()

	// Get the first day of the current month
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Get the first day of the next month
	firstDayNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	// Format the dates as YYYY-MM-DD
	firstDayString := firstDayOfMonth.Format("2006-01-02")
	lastDayString := firstDayNextMonth.Format("2006-01-02")

	return firstDayString, lastDayString
}
