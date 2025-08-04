package models

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var DB *sql.DB

type DBConfig struct {
	Port     int    `env:"DB_PORT"`
	Host     string `env:"DB_HOST"`
	Password string `env:"DB_PASSWORD"`
	User     string `env:"DB_USER"`
	Database string `env:"DB_DATABASE"`
}

func InitDatabase(config DBConfig) (*sql.DB, error) {

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

	return DB, nil

}

func CloseDatabase() error {
	if DB != nil {
		fmt.Println("Closing the database connection...")
		return DB.Close()
	}
	return nil
}
