package controllers

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

func CashController(ctx *gin.Context) {
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
	err := db.DB.QueryRow(context.Background(), "SELECT name FROM clients WHERE api_key = $1", body.APIKey).Scan(&client)
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
	_, err = db.DB.Exec(context.Background(), sqlStatement, _uuid, time.Now(), time.Now(), client, body.Contact, body.Amount, body.Detail, body.Note)
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

func CashesWihtAuth(ctx *gin.Context) {
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

	sqlStatement := `SELECT c.uuid, c.amount, c.contact, c.client, c.detail, c.note, c.created_at FROM cashes c`
	sqlFilters := ""
	if len(queries) > 0 {
		sqlFilters += " WHERE "
		sqlFilters += strings.Join(queries, " AND ")
	}
	sqlStatement += sqlFilters
	sqlStatement += " ORDER BY created_at DESC "
	sqlStatement += fmt.Sprintf(" offset $%v limit $%v", index+1, index+2)
	rows, err := db.DB.Query(context.Background(), sqlStatement, valuesWithPagination...)
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
		err := rows.Scan(&cash.UUID, &cash.Amount, &cash.Contact, &cash.Client, &cash.Detail, &cash.Note, &cash.CreatedAt)
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
	err = db.DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM cashes"+sqlFilters, values...).Scan(&totalCashes)
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

func CashesByUUID(ctx *gin.Context) {
	// Get UUID from URL param
	uuid, ok := ctx.Params.Get("uuid")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "uuid is required",
			"message": "Coulnd't find UUID",
		})
		return
	}

	// Find the cash with given UUID
	var cash models.CashBodyResponse
	err := db.DB.QueryRow(context.Background(), "SELECT uuid, created_at, client, contact, amount, detail, note FROM cashes where uuid = $1", uuid).Scan(&cash.UUID, &cash.CreatedAt, &cash.Client, &cash.Contact, &cash.Amount, &cash.Detail, &cash.Note)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   err.Error(),
			"message": "Couldn't find the cash details",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"cash": cash,
	})
}
