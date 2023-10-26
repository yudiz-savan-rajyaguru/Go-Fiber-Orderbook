package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/opinion-trading/helper"
	"github.com/opinion-trading/helper/response"
	redisclient "github.com/opinion-trading/redis_client"
)

var ctx = context.Background()

// func BestAvailableMatch(c *fiber.Ctx) error {
// 	var body TradeRequestBody

// 	err := c.BodyParser(&body)
// 	if err != nil {
// 		return helper.SendResponse(c, helper.Response{
// 			Code:    fiber.StatusBadRequest,
// 			Message: err.Error(),
// 		})
// 	}
// 	// handle enum
// 	m := make(map[string]string)
// 	m["side"] = body.Side.Valid()
// 	m["flag"] = body.Flag.Valid()

// 	if err := helper.ValidateEnum(m); len(err) != 0 {
// 		return helper.SendResponse(c, helper.Response{
// 			Code:    fiber.StatusBadRequest,
// 			Message: "Enum validation failed",
// 			Data:    err,
// 		})
// 	}

// 	// get different side data with matching qty and price 10 - current
// 	getPrice := 10 - body.Price
// 	var getSide string
// 	if body.Side == "YES" {
// 		getSide = "NO"
// 	} else {
// 		getSide = "YES"
// 	}
// 	opponentKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, getSide, getPrice, body.Qty)

// 	list, err := redisclient.Rdb.LRange(ctx, opponentKey, 0, -1).Result()

// 	if err != nil {
// 		return helper.ErrorHandler(c, err)
// 	}

// 	jsonString, _ := helper.Marshal(&body)

// 	// when first trade happen
// 	if len(list) == 0 {
// 		// push order to the redis list
// 		redisKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, body.Side, body.Price, body.Qty)

// 		_, err = redisclient.Rdb.RPush(ctx, redisKey, jsonString).Result()
// 		if err != nil {
// 			return helper.ErrorHandler(c, err)
// 		}
// 		return helper.SendResponse(c, helper.Response{
// 			Code:    fiber.StatusOK,
// 			Message: fmt.Sprintf("%s %s", response.Words["order"], response.Message["not_match"]),
// 		})
// 	}

// 	// when the opponent data is present
// 	var opponentList []TradeRequestBody
// 	for _, val := range list {
// 		var opponentStruct TradeRequestBody

// 		if err := json.Unmarshal([]byte(val), &opponentStruct); err != nil {
// 			return helper.ErrorHandler(c, err)
// 		}

// 		// matching condition
// 		if opponentStruct.Qty == body.Qty && (opponentStruct.Price+body.Price == 10) {
// 			// perform redis transaction
// 			var orderbook Orderbook
// 			orderbook.EID = body.EID
// 			orderbook.CreatedAt = time.Now()
// 			if body.Side == "YES" {
// 				orderbook.YES_UID = body.UID
// 				orderbook.YES_Price = body.Price
// 				orderbook.YES_Qty = body.Qty
// 				orderbook.NO_UID = opponentStruct.UID
// 				orderbook.NO_Price = opponentStruct.Price
// 				orderbook.NO_Qty = opponentStruct.Qty
// 			} else if body.Side == "NO" {
// 				orderbook.YES_UID = opponentStruct.UID
// 				orderbook.YES_Price = opponentStruct.Price
// 				orderbook.YES_Qty = opponentStruct.Qty
// 				orderbook.NO_UID = body.UID
// 				orderbook.NO_Price = body.Price
// 				orderbook.NO_Qty = body.Qty
// 			}

// 			jsonData, _ := helper.Marshal(&orderbook)
// 			_, err = redisclient.Rdb.RPush(ctx, "orderbook", jsonData).Result()
// 			if err != nil {
// 				return helper.ErrorHandler(c, err)
// 			}

// 			jsonData, _ = helper.Marshal(&opponentStruct)
// 			_, err = redisclient.Rdb.LRem(ctx, opponentKey, 1, jsonData).Result()
// 			if err != nil {
// 				return helper.ErrorHandler(c, err)
// 			}
// 			opponentList = append(opponentList, opponentStruct)
// 			return helper.SendResponse(c, helper.Response{
// 				Code:    fiber.StatusOK,
// 				Message: response.Message["match_found"],
// 				Data:    opponentList,
// 			})
// 		}
// 	}

// 	// else create data entry in redis list
// 	redisKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, body.Side, body.Price, body.Qty)

// 	_, err = redisclient.Rdb.RPush(ctx, redisKey, jsonString).Result()
// 	if err != nil {
// 		return helper.ErrorHandler(c, err)
// 	}
// 	return helper.SendResponse(c, helper.Response{
// 		Code:    fiber.StatusOK,
// 		Message: fmt.Sprintf("%s %s", response.Words["order"], response.Message["not_match"]),
// 	})
// }

func BestAvailableMatch(c *fiber.Ctx) error {
	var body TradeRequestBody

	err := c.BodyParser(&body)
	if err != nil {
		return helper.ErrorHandler(c, err)
	}
	// handle enum
	m := make(map[string]string)
	m["side"] = body.Side.Valid()
	m["flag"] = body.Flag.Valid()

	if err := helper.ValidateEnum(m); len(err) != 0 {
		return helper.SendResponse(c, helper.Response{
			Code:    fiber.StatusBadRequest,
			Message: "Enum validation failed",
			Data:    err,
		})
	}
	getPrice := 10 - body.Price
	var getSide string
	if body.Side == "YES" {
		getSide = "NO"
	} else {
		getSide = "YES"
	}

	opponentKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, getSide, getPrice, body.Qty)
	redisKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, body.Side, body.Price, body.Qty)

	// first find to the opponentKey list
	data, err := redisclient.Rdb.LPop(ctx, opponentKey).Result()
	if err != nil {
		// if opponentKey list not found than it's a new order push to the list
		jsonData, _ := helper.Marshal(&body)
		_, err := redisclient.Rdb.RPush(ctx, redisKey, jsonData).Result()
		if err != nil {
			return helper.ErrorHandler(c, err)
		}
		return helper.SendResponse(c, helper.Response{
			Code:    fiber.StatusOK,
			Message: fmt.Sprintf("%s %s", response.Words["order"], response.Message["not_match"]),
		})
	}

	// start transaction for the matching insert records
	_, err = redisclient.Rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		var opponentStruct TradeRequestBody
		// convert string to the struct
		if err := json.Unmarshal([]byte(data), &opponentStruct); err != nil {
			return helper.ErrorHandler(c, err)
		}
		var orderbook Orderbook
		orderbook.EID = body.EID
		orderbook.CreatedAt = time.Now()
		if body.Side == "YES" {
			orderbook.YES_UID = body.UID
			orderbook.YES_Price = body.Price
			orderbook.YES_Qty = body.Qty
			orderbook.NO_UID = opponentStruct.UID
			orderbook.NO_Price = opponentStruct.Price
			orderbook.NO_Qty = opponentStruct.Qty
		} else if body.Side == "NO" {
			orderbook.YES_UID = opponentStruct.UID
			orderbook.YES_Price = opponentStruct.Price
			orderbook.YES_Qty = opponentStruct.Qty
			orderbook.NO_UID = body.UID
			orderbook.NO_Price = body.Price
			orderbook.NO_Qty = body.Qty
		}
		orderBookData, err := json.Marshal(orderbook)
		if err != nil {
			p.Discard()
		}
		p.RPush(ctx, "orderbook", orderBookData)
		opponentStructData, err := json.Marshal(orderbook)
		if err != nil {
			p.Discard()
		}
		p.LRem(ctx, opponentKey, 1, opponentStructData)

		_, err = p.Exec(ctx)
		return err
	})

	if err != nil {
		return helper.ErrorHandler(c, err)
	}

	return helper.SendResponse(c, helper.Response{
		Code:    fiber.StatusOK,
		Message: response.Message["match_found"],
		Data:    err,
	})
}
