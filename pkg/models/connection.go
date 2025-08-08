package models

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gqvz/mvc/pkg/config"
	"time"
)

var DB *sql.DB

func InitDatabase(config config.DBConfig) (*sql.DB, error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.Database)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening the database: %v", err)
	}

	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	err = DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	driver, _ := mysql.WithInstance(DB, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance("file://database/migrations", "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("error creating migration instance: %v", err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("error applying migrations: %v", err)
	}
	fmt.Println("Database connection established successfully")
	return DB, nil

}

func CloseDatabase() error {
	if DB != nil {
		fmt.Println("Closing the database connection...")
		return DB.Close()
	}
	return nil
}
