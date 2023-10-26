package trade

import (
	"github.com/gofiber/fiber/v2"
	h "github.com/opinion-trading/helper"
)

func TradRoute(route fiber.Router) {
	routes := route.Group("/trade")

	routes.Post("", h.BodyValidator(&TradeRequestBody{}), BestAvailableMatch)
	// routes.Put("", h.)
}
