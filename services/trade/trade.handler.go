package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/opinion-trading/helper"
	redisclient "github.com/opinion-trading/redis_client"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func matchOrders(ctx context.Context, body *TradeRequestBody, flag string, preRedisKey string, newPrice float32, newQty int) MatchResult {
	getPrice := 10 - body.Price
	var getSide string
	if body.Side == "YES" {
		getSide = "NO"
	} else {
		getSide = "YES"
	}
	var opponentKey string
	if flag == "Update" {
		getPrice = 10 - newPrice
		opponentKey = fmt.Sprintf("%s:%s:%g:%d", body.EID, getSide, getPrice, newQty)
	} else {
		opponentKey = fmt.Sprintf("%s:%s:%g:%d", body.EID, getSide, getPrice, body.Qty)
	}
	redisKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, body.Side, body.Price, body.Qty)
	newKeyForUpdate := fmt.Sprintf("%s:%s:%g:%d", body.EID, body.Side, newPrice, newQty)
	// fmt.Printf("UpdateKey>> %s, redisKey>> %s, opponentKey>> %s\n", newKeyForUpdate, redisKey, opponentKey)
	// First, find the opponentKey list
	data, err := redisclient.Rdb.LPop(ctx, opponentKey).Result()
	// fmt.Println("Data>> Err>>", data, err)
	if err != nil {
		// If opponentKey list is not found, it's a new order, so push it to the list
		if err == redis.Nil {
			if flag == "Update" {
				jsonString, err := json.Marshal(body)
				if err != nil {
					return MatchResult{
						Error: err,
						Msg:   "Update json Marshal error",
					}
				}
				_, err = redisclient.Rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
					// 1. delete entry from the previous list
					p.LRem(ctx, preRedisKey, 1, jsonString)

					// 2. make a new list for the new entry
					body.Price = newPrice
					body.Qty = newQty
					newData, err := json.Marshal(body)
					if err != nil {
						p.Discard()
					}
					p.RPush(ctx, newKeyForUpdate, newData)
					_, err = p.Exec(ctx)
					return err
				})
				if err != nil {
					return MatchResult{
						Data:  nil,
						Error: err,
						Msg:   "Something went wrong in update",
					}
				}
				return MatchResult{
					Code:  fiber.StatusOK,
					Data:  body,
					Error: nil,
					Msg:   "Updated Order does not match",
				}
			} else {
				jsonData, _ := helper.Marshal(&body)
				_, err := redisclient.Rdb.RPush(ctx, redisKey, jsonData).Result()
				if err != nil {
					return MatchResult{
						Data:  nil,
						Error: err,
						Msg:   "Something went wrong in the matching",
					}
				}
				return MatchResult{
					Code:  fiber.StatusOK,
					Data:  body,
					Error: nil,
					Msg:   "Order does not match",
				}
			}
		}
		return MatchResult{
			Data:  nil,
			Error: err,
			Msg:   "Something went wrong in the matching",
		}
	}

	// Start a transaction for the matching and insert records
	var orderbook Orderbook
	var opponentStruct TradeRequestBody
	// Convert string to the struct
	if err := json.Unmarshal([]byte(data), &opponentStruct); err != nil {
		return MatchResult{
			Error: err,
			Msg:   "Orderbook Unmarshal error",
		}
	}

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

	// matching transaction
	_, err = redisclient.Rdb.TxPipelined(ctx, func(p redis.Pipeliner) error {
		orderBookData, err := json.Marshal(orderbook)
		if err != nil {
			p.Discard()
		}
		p.RPush(ctx, "orderbook", orderBookData)

		opponentStructData, err := json.Marshal(opponentStruct)
		if err != nil {
			p.Discard()
		}
		p.LRem(ctx, opponentKey, 1, opponentStructData)

		// only if update operation is perform
		if flag == "Update" {
			jsonString, err := json.Marshal(body)
			// 1. delete entry from the previous list
			if err != nil {
				p.Discard()
			}
			p.LRem(ctx, preRedisKey, 1, jsonString)
		}

		_, err = p.Exec(ctx)
		return err
	})

	if err != nil {
		return MatchResult{
			Data:  nil,
			Error: err,
			Msg:   "Something went wrong in the matching",
		}
	}
	return MatchResult{
		Code:      fiber.StatusOK,
		Orderbook: &orderbook,
		Error:     nil,
		Msg:       "Match found successful",
	}
}

func PlaceOrder(c *fiber.Ctx) error {
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

	result := matchOrders(ctx, &body, "", "", 0, 0)

	if result.Error != nil {
		return helper.ErrorHandler(c, result.Error)
	}

	if result.Orderbook != nil {
		return helper.SendResponse(c, helper.Response{
			Code:    int(result.Code),
			Message: result.Msg,
			Data:    result.Orderbook,
		})
	}
	return helper.SendResponse(c, helper.Response{
		Code:    int(result.Code),
		Message: result.Msg,
		Data:    result.Data,
	})
}

func UpdateOrder(c *fiber.Ctx) error {
	var body UpdateTradeRequestBody

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

	// if the new price and qty is same
	if body.NewPrice == body.PrePrice && body.NewQty == body.PreQty {
		return helper.SendResponse(c, helper.Response{
			Code:    fiber.StatusBadRequest,
			Message: "New Price and new qty must not be same",
			Data:    nil,
		})
	}

	preRedisKey := fmt.Sprintf("%s:%s:%g:%d", body.EID, body.Side, body.PrePrice, body.PreQty)

	var deleteData TradeRequestBody
	deleteData.UID = body.UID
	deleteData.EID = body.EID
	deleteData.Price = body.PrePrice
	deleteData.Qty = body.PreQty
	deleteData.Side = body.Side
	deleteData.Flag = body.Flag

	// check if exist or not
	data, err := json.Marshal(deleteData)
	if err != nil {
		return helper.ErrorHandler(c, err)
	}
	// if record not found in the list
	_, err = redisclient.Rdb.LPos(ctx, preRedisKey, string(data), redis.LPosArgs{Rank: 1}).Result()
	if err != nil {
		return helper.SendResponse(c, helper.Response{
			Code:    fiber.StatusBadRequest,
			Message: "Update order not found in order list",
			Data:    nil,
		})
	}

	// matching the new order
	result := matchOrders(ctx, &deleteData, "Update", preRedisKey, body.NewPrice, body.NewQty)
	if result.Error != nil {
		return helper.ErrorHandler(c, result.Error)
	}

	if result.Orderbook != nil {
		return helper.SendResponse(c, helper.Response{
			Code:    int(result.Code),
			Message: result.Msg,
			Data:    result.Orderbook,
		})
	}
	return helper.SendResponse(c, helper.Response{
		Code:    int(result.Code),
		Message: result.Msg,
		Data:    result.Data,
	})
}
