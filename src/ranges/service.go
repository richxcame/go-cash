package ranges

import (
	"context"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5"
)

func GetRangesService(offset, limit int) ([]models.RangeBodyResponse, int, string, string) {

	// Find ranges
	var rangeBodies []models.RangeBodyResponse
	rows, err := db.DB.Query(context.Background(), slRngsByLntSql, offset, limit)
	if err != nil {
		return nil, 0, err.Error(), "Something went wrong"
	}
	for rows.Next() {
		var rangeBody models.RangeBodyResponse
		err := rows.Scan(&rangeBody.UUID, &rangeBody.CreatedAt, &rangeBody.UpdatedAt, &rangeBody.Client, &rangeBody.Detail, &rangeBody.Note)
		if err != nil {
			logger.Errorf("Scan error %v", err)
			return nil, 0, err.Error(), "Scan error"

		}
		rangeBodies = append(rangeBodies, rangeBody)
	}

	resultRanges := make([]models.RangeBodyResponse, 0)
	for _, v := range rangeBodies {
		var rangeBody models.RangeBodyResponse
		err := db.DB.QueryRow(context.Background(), slRngByClntSql, v.CreatedAt, v.Client).Scan(&rangeBody.CreatedAt, &rangeBody.Client)
		if err != nil && err != pgx.ErrNoRows {
			return nil, 0, err.Error(), "Couldn't find searching range"
		} else if err == pgx.ErrNoRows {
			rangeBody.CreatedAt = time.Date(2001, 12, 28, 0, 0, 0, 0, time.Now().Location())
		}

		// Search sum of cashes between the two ranges
		var totalAmount *float64
		err = db.DB.QueryRow(context.Background(), slSumCshsByClntSql, rangeBody.CreatedAt, v.CreatedAt, v.Client).Scan(&totalAmount)
		if err != nil {
			return nil, 0, err.Error(), "Couldn't find total amount of the range"
		}
		if totalAmount == nil {
			defaultTotalAmount := float64(0)
			totalAmount = &defaultTotalAmount
		}

		// Search sum and amount of one currency cashes between the two ranges
		var currencies models.Currencies
		for _, vCurrency := range []uint{1, 5, 10, 20, 50, 100} {
			var currency models.Currency
			rowOne, err := db.DB.Query(context.Background(), slSumCshByGrpSql, rangeBody.CreatedAt, v.CreatedAt, v.Client, vCurrency)
			if err != nil {
				return nil, 0, err.Error(), "Couldn't find total amount of the range"
			}

			for rowOne.Next() {
				err := rowOne.Scan(&currency.TotalAmount, &currency.Amount)
				if err != nil {
					logger.Errorf("Scan error %v", err)
					return nil, 0, err.Error(), "Scan error"
				}
			}

			switch vCurrency {
			case 1:
				currencies.One = currency
			case 5:
				currencies.Five = currency
			case 10:
				currencies.Ten = currency
			case 20:
				currencies.Twenty = currency
			case 50:
				currencies.Fifty = currency
			case 100:
				currencies.OneHundred = currency
			}
			// it's need rowOne.Close()
		}

		resultRanges = append(resultRanges, models.RangeBodyResponse{
			UUID:        v.UUID,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			Client:      v.Client,
			Note:        v.Note,
			Detail:      v.Detail,
			TotalAmount: *totalAmount,
			Currencies:  currencies,
		})
	}

	// Find total count of ranges
	totalRanges := 0
	err = db.DB.QueryRow(context.Background(), slAllRngsSql).Scan(&totalRanges)
	if err != nil {
		return nil, 0, err.Error(), "Couldn't find total number of ranges"
	}
	return rangeBodies, totalRanges, "", ""

}
