package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/opinion-trading/helper"
	"github.com/opinion-trading/services/order"
	"github.com/opinion-trading/services/trade"
)

func registerV1Routes(apiRoute fiber.Router) {
	v1Route := apiRoute.Group("/v1")

	trade.TradRoute(v1Route)
	order.OrderRoute(v1Route)
}

func RegisterRoutes(app *fiber.App) {
	apiRoute := app.Group("/api")

	registerV1Routes(apiRoute)

	app.Get("/health-check", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(helper.Response{
			Code:    fiber.StatusOK,
			Message: "Server is running",
		})
	})

	app.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(helper.Response{
			Code:    fiber.StatusNotFound,
			Message: "Route Not Found!!",
		})
	})
}
