package main

import (
	"gocash/config"
	"gocash/pkg/db"
	"gocash/routes"
)

func init() {
	config.InitGlobals()
	db.CreateDB()
}

func main() {
	routes.InitRoutes().Run()
}
