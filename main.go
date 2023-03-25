package main

import (
	"context"
	"fmt"
	"gocash/pkg/arrs"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

type CashBody struct {
	APIKey  string  `json:"api_key" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Contact string  `json:"contact" binding:"required"`
	Detail  string  `json:"detail"`
	Note    string  `json:"note"`
}

type RangeBody struct {
	APIKey string `json:"api_key" binding:"required"`
	Detail string `json:"detail"`
	Note   string `json:"note"`
}

type RangeBodyResponse struct {
	Detail    string    `json:"detail"`
	Note      string    `json:"note"`
	Client    string    `json:"client"`
	CreatedAt time.Time `json:"created_at"`
}

type CashBodyResponse struct {
	Amount    float64   `json:"amount" binding:"required"`
	Detail    string    `json:"detail"`
	Note      string    `json:"note"`
	Client    string    `json:"client"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Database instance
	db := db.CreateDB()
	defer db.Close()

	r := gin.Default()

	r.POST("/cashes", func(ctx *gin.Context) {
		// Get request body
		var body CashBody
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
		err := db.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
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
		sqlStatement := `
		INSERT INTO cashes (uuid, created_at, updated_at, client, contact, amount, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = db.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Contact, body.Amount, body.Detail, body.Note)
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
	})

	r.POST("/ranges", func(ctx *gin.Context) {
		// Get request body
		var body RangeBody
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
		err := db.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
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
		sqlStatement := `
		INSERT INTO ranges (uuid, created_at, updated_at, client, detail, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = db.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Detail, body.Note)
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
	})

	r.GET("/ranges", func(ctx *gin.Context) {
		offsetQuery := ctx.DefaultQuery("offset", "0")
		limitQuery := ctx.DefaultQuery("limit", "20")
		offset, err := strconv.Atoi(offsetQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "offset value must be convertable to integer",
			})
			return
		}
		limit, err := strconv.Atoi(limitQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "limit value must be convertable to integer",
			})
			return
		}

		var rangeBodies []RangeBodyResponse
		sqlStatement := `SELECT r.detail, r.note, r.client, r.created_at, r.updated_at FROM ranges r ORDER BY r.created_at DESC OFFSET $1 LIMIT $2;`
		rows, err := db.Query(context.Background(), sqlStatement, offset, limit)
		if err != nil {
			logger.Errorf("ranges search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Something went wrong",
			})
			return
		}
		for rows.Next() {
			var rangeBody RangeBodyResponse
			err := rows.Scan(&rangeBody.Detail, &rangeBody.Note, &rangeBody.Client, &rangeBody.CreatedAt)
			if err != nil {
				logger.Errorf("Scan error %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err,
					"message": "Scan error",
				})
				return
			}
			rangeBodies = append(rangeBodies, rangeBody)
		}
		ctx.JSON(200, gin.H{
			"message": "Successfully get ranges",
			"ranges":  rangeBodies,
		})
	})

	// TODO: need fix
	r.GET("/reports", func(ctx *gin.Context) {
		var cashBody CashBodyResponse
		var cashBodies []CashBodyResponse
		sqlStatement := `
		SELECT
		MAX(client),
		MAX(contact),
		SUM(amount),
		MAX(detail),
		MAX(created_at)
	FROM
		cashes
	WHERE
		created_at > $1
		AND created_at < $2
	GROUP BY
		contact
	ORDER BY
		MAX(created_at) DESC;`
		rows, err := db.Query(context.Background(), sqlStatement, "2023-03-19", "2023-03-20")
		if err != nil {
			logger.Errorf("cashes search error %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err,
				"message": "Something went wrong",
			})
			return
		}
		for rows.Next() {
			err := rows.Scan(&cashBody.Client, &cashBody.Contact, &cashBody.Amount, &cashBody.Detail, &cashBody.CreatedAt)
			if err != nil {
				logger.Errorf("Scan error %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   err,
					"message": "Scan error",
				})
				return
			}
			cashBodies = append(cashBodies, cashBody)
		}
		ctx.JSON(200, gin.H{
			"message": "Successfully get cashes",
			"ranges":  cashBodies,
		})
	})

	// /cashes
	// Filters: amount, detail, note, client, contact as array
	// Pagination: offset, limit with defaults respectively 0, 50
	r.GET("/cashes", func(ctx *gin.Context) {
		offsetQuery := ctx.DefaultQuery("offset", "0")
		limitQuery := ctx.DefaultQuery("limit", "50")
		offset, err := strconv.Atoi(offsetQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "offset value must be convertable to integer",
			})
			return
		}
		limit, err := strconv.Atoi(limitQuery)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "limit value must be convertable to integer",
			})
			return
		}

		urlQueries := ctx.Request.URL.Query()
		index := 0
		var values []interface{}
		var queries []string
		for k, v := range urlQueries {
			if arrs.Contains([]string{"uuid", "client", "contact", "amount", "detail", "note"}, k) {
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

		sqlStatement := `SELECT c.amount, c.contact, c.client, c.detail, c.note, c.created_at FROM cashes c`
		sqlFilters := ""
		if len(queries) > 0 {
			sqlFilters += " WHERE "
			sqlFilters += strings.Join(queries, " AND ")
		}
		sqlStatement += sqlFilters
		sqlStatement += " ORDER BY created_at DESC "
		sqlStatement += fmt.Sprintf(" offset $%v limit $%v", index+1, index+2)
		rows, err := db.Query(context.Background(), sqlStatement, valuesWithPagination...)
		if err != nil {
			logger.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Couldn't search from cashes",
			})
			return
		}
		defer rows.Close()

		cashes := make([]CashBodyResponse, 0)
		for rows.Next() {
			var cash CashBodyResponse
			err := rows.Scan(&cash.Amount, &cash.Contact, &cash.Client, &cash.Detail, &cash.Note, &cash.CreatedAt)
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
		err = db.QueryRow(context.Background(), "SELECT COUNT(*) FROM cashes"+sqlFilters, values...).Scan(&totalCashes)
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
	})

	r.Run()
}
