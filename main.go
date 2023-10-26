package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/helmet/v2"
	"github.com/opinion-trading/config"
	"github.com/opinion-trading/database"
	"github.com/opinion-trading/helper/response"
	"github.com/opinion-trading/middleware"
	"github.com/opinion-trading/migration"
	redisclient "github.com/opinion-trading/redis_client"
)

func init() {
	config.InitEnvVariables()
	response.ResponseMsg()
	response.ResponseWord()
	database.ConnectDB()
	migration.LoadAllSchema()
	redisclient.Connect()
	// cron.ImplementCron()
}

func main() {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
	})
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(helmet.New())
	app.Use(middleware.RecoveryMiddleware)

	middleware.RegisterRoutes(app)
	// For the hot reload script
	// nodemon --watch './**/*.go' --signal SIGTERM --exec 'go' run ./main.go
	log.Println("Server running on...", config.ConfigEnv.PORT)
	log.Fatal(app.Listen(":" + config.ConfigEnv.PORT).Error())
}
