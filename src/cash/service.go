package cash

import (
	"context"
	"gocash/models"
	"gocash/pkg/db"
	"gocash/pkg/logger"
)

func GetDatesCashesService(ctx *context.Context, date1, date2 string) (*models.ResultDatesCashes, string, error) {
	var sumCh float64
	var sumT models.Toleg
	var result models.ResultDatesCashes
	err := db.DB.QueryRow(*ctx, slDatesCashSql, date1, date2).Scan(&sumCh)
	if err != nil && err.Error() != "can't scan into dest[0]: cannot scan NULL into *float64" {
		logger.Errorf("get Dates cashes error %v", err)
		return nil, "get Dates cashes error", err
	}

	err = db.GoToleg.QueryRow(*ctx, slDatesTolegSql, date1, date2).Scan(&sumT.Reasons, &sumT.TotalSum)
	if err != nil && err.Error() != "can't scan into dest[0]: cannot scan NULL into *float64" {

		logger.Errorf("get Dates transactions error %v", err)
		return nil, "get Dates transactions error", err
	}

	result.TotalToleg = sumT
	result.TotalCash = sumCh
	return &result, "", nil
}
