package database

import (
	"log"
	"time"

	"github.com/opinion-trading/config"
	"github.com/opinion-trading/helper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	db, err := gorm.Open(mysql.Open(config.ConfigEnv.SQL_DB_URL), &gorm.Config{
		// Logger: logger.Default,
		// Logger: logger.Default.LogMode(logger.Info),
		// QueryFields: true,
		// 	SkipInitializeWithVersion: false, // auto configure based on currently MySQL
	})

	if err != nil {
		log.Panic("DB connection fail", err)
	} else {
		DB = db

		DB, err := db.DB()
		if err != nil {
			log.Fatal("Database connection error ::", err)
		}
		DB.SetConnMaxIdleTime(time.Duration(config.ConfigEnv.SQL_MAX_IDLE_CONNECTION))
		DB.SetConnMaxLifetime(time.Duration(config.ConfigEnv.SQL_MAX_OPEN_CONNECTION))

		log.Print("DB connected success...")
		dbName := helper.ExtractDatabaseName(config.ConfigEnv.SQL_DB_URL)
		log.Print("Database Name: ", dbName)
	}
}
