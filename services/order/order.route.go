package order

import (
	"github.com/gofiber/fiber/v2"
	"github.com/opinion-trading/helper"
)

func OrderRoute(route fiber.Router) {
	routes := route.Group("/order")

	routes.Post("", helper.BodyValidator(&RequestBody{}), MakeOrderWithRedis)
	routes.Put("", helper.BodyValidator(&UpdateBody{}), UpdateOrderQtyWithRedis)
}
