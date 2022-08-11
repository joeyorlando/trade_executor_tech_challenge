package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joeyorlando/trade_executor_tech_challenge/internal/binance"
	_ "github.com/mattn/go-sqlite3"
)

const driverName = "sqlite3"

type Database struct {
	ConnectionPool      *sql.DB
	DatabaseName        string
	MigrationsDirectory string
}

func NewDatabase(databaseFilePath, databaseName, migrationsDirectory string) (Database, error) {
	db, err := sql.Open(driverName, databaseFilePath)
	return Database{
		ConnectionPool:      db,
		DatabaseName:        databaseName,
		MigrationsDirectory: migrationsDirectory,
	}, err
}

func (db *Database) RunMigrations() error {
	driver, err := sqlite3.WithInstance(db.ConnectionPool, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("invalid target sqlite instance, %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", db.MigrationsDirectory),
		db.DatabaseName, driver)
	if err != nil {
		return fmt.Errorf("error running the database migrations, %w", err)
	}

	return m.Up()
}

func (db *Database) PersistFulfilledOrder(order binance.LimitOrder, orderSplits []binance.OrderSplit) error {
	var orderId int
	dbConn := db.ConnectionPool

	tx, err := dbConn.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO fulfilled_orders(
			symbol,
			quantity,
			price
		) VALUES(
			?,
			?,
			?
		)
		RETURNING id;
	`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	// create the order
	if err = stmt.QueryRow(order.Symbol, order.Quantity, order.Price).Scan(&orderId); err != nil {
		return err
	}

	// create the order splits
	for _, orderSplit := range orderSplits {
		stmt, err = tx.Prepare(`
			INSERT INTO fulfilled_order_splits(
				order_id,
				update_id,
				quantity,
				price
			) VALUES(
				?,
				?,
				?,
				?
			);
		`)

		_, err = stmt.Exec(orderId, orderSplit.UpdateId, orderSplit.BidQuantity, orderSplit.BidPrice)
		if err != nil {
			return err
		}

		defer stmt.Close()
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
