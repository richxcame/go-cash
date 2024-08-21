package db

import (
	"context"
	"gocash/pkg/logger"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool
var GoToleg *pgxpool.Pool

func CreateDB() {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatalf("Failed to create new pool %v", err)
	}
	if err := dbpool.Ping(context.Background()); err != nil {
		logger.Fatalf("Couldn't connect to database %v", err)
	}

	DB = dbpool

	dbpool, err = pgxpool.New(context.Background(), os.Getenv("GOTOLEG_DATABASE_URL"))
	if err != nil {
		logger.Fatalf("Failed to create new gotolegdb pool %v", err)
	}
	if err := dbpool.Ping(context.Background()); err != nil {
		logger.Fatalf("Couldn't connect to gotoleg database %v", err)
	}

	GoToleg = dbpool

}
