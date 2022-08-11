package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joeyorlando/trade_executor_tech_challenge/cmd/server"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/binance"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/database"
)

func main() {
	orderTimeoutSeconds, err := strconv.Atoi(os.Getenv("ORDER_TIMEOUT_SECONDS"))
	if err != nil {
		log.Fatal("ORDER_TIMEOUT_SECONDS must be an integer")
	}

	bin := binance.NewBinance(orderTimeoutSeconds)
	dbMigrationsFilePath, _ := filepath.Abs("./migrations")
	databaseFilePath, _ := filepath.Abs("./database.db")

	db, err := database.NewDatabase(databaseFilePath, "tech_challenge", dbMigrationsFilePath)
	if err != nil {
		log.Fatal(err)
	}

	db.RunMigrations()

	server := server.NewServer(os.Getenv("HTTP_PORT"), bin, db)
	server.Run()
}
