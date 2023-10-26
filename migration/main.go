package migration

import (
	"log"

	"github.com/opinion-trading/database"
	"github.com/opinion-trading/services/cron"
)

func Migration(models ...interface{}) {
	for _, model := range models {
		err := database.DB.AutoMigrate(&model)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func LoadAllSchema() {
	Migration(
		&cron.OrderBookModel{},
	)
	log.Print("Schema migration success...")
	// func() {
	// 	database.DB.Migrator().DropColumn(&auth.UserModel{}, "company_id")
	// }()
}
