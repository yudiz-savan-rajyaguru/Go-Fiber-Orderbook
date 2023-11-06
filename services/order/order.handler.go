package order

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	h "github.com/opinion-trading/helper"
	redisclient "github.com/opinion-trading/redis_client"
	"github.com/redis/go-redis/v9"
)

var bidsOfYes []Order
var bidsOfNo []Order
var orderbook []PlacedOrder

var ctx = context.Background()

func findBestMatch(side string, price float32, userId string, qty int16, eventId string) int16 {
	var placeOrder PlacedOrder
	placeOrder.CreatedAt = time.Now()
	placeOrder.EventId = eventId
	newPrice := 10 - price
	// 1. find the matching order to the no list
	if side == "YES" {
		placeOrder.YesUserId = userId
		placeOrder.YesQty = qty
		placeOrder.YesPrice = price
		// Optimization require for this like Binary search
		for index, val := range bidsOfYes {
			if val.Price == newPrice && val.Qty == qty {
				placeOrder.NoPrice = val.Price
				placeOrder.NoQty = val.Qty
				placeOrder.NoUserId = val.UserId
				orderbook = append(orderbook, placeOrder)
				// delete the match data form the bids
				bidsOfYes = append(bidsOfYes[:index], bidsOfYes[index+1:]...)

				// update both user balance of yes side and no side
				return 1
			}
		}
		return 0
	} else {
		placeOrder.NoUserId = userId
		placeOrder.NoQty = qty
		placeOrder.NoPrice = price
		// Optimization require for this using Binary search
		for index, val := range bidsOfYes {
			if val.Price == newPrice && val.Qty == qty && val.UserId != userId {
				placeOrder.YesPrice = val.Price
				placeOrder.YesQty = val.Qty
				placeOrder.YesUserId = val.UserId
				orderbook = append(orderbook, placeOrder)
				// delete the match data form the bids
				bidsOfNo = append(bidsOfNo[:index], bidsOfNo[index+1:]...)

				// update both user balance of yes side and no side
				return 1
			}
		}
		return 0
	}
}

// func findMatchWithRedis(side string, price float32, userId string, qty int16, eventId string) int16 {
// 	var placeOrder PlacedOrder
// 	placeOrder.CreatedAt = time.Now()
// 	placeOrder.EventId = eventId
// 	// newPrice := 10 - price
// 	// 1. find the matching order to the no list
// 	if side == "YES" {
// 		placeOrder.YesUserId = userId
// 		placeOrder.YesQty = qty
// 		placeOrder.YesPrice = price
// 		redisKey := fmt.Sprintf("bidsOfYes:%s", eventId)
// 		query := fmt.Sprintf("bidsOfYes:%s*->Price bidsOfYes:%s*->Qty", eventId, eventId)
// 		sortResults, err := redisClient.Rdb.Sort(ctx, redisKey, &redis.Sort{By: query}).Result()
// 		fmt.Println("Err in sort>>", err)
// 		for _, result := range sortResults {
// 			fmt.Println(result)
// 		}
// 	} else {

// 	}
// 	return 1
// }

func MakeOrder(c *fiber.Ctx) error {
	var body RequestBody

	err := c.BodyParser(&body)
	if err != nil {
		return h.ErrorHandler(c, err)
	}
	// handle enum
	m := make(map[string]string)
	m["side"] = body.Side.Valid()
	m["purchaseFlag"] = body.PurchaseFlag.Valid()

	if err := h.ValidateEnum(m); len(err) != 0 {
		return h.SendResponse(c, h.Response{
			Code:    fiber.StatusBadRequest,
			Message: "Enum validation failed",
			Data:    err,
		})
	}
	// 1. match the qty for the order
	result := findBestMatch(string(body.Side), body.Price, body.UserID, body.Qty, body.EventID)

	if result == 1 {
		return h.SendResponse(c, h.Response{
			Code:    fiber.StatusOK,
			Message: "Match Found successful",
			Data:    orderbook,
		})
	}

	// 2. if match not found than order place into the bids array according to the side
	var order Order
	order.Price = body.Price
	order.Qty = body.Qty
	order.UserId = body.UserID
	var wg sync.WaitGroup
	if body.Side == "YES" {
		wg.Add(2)
		go func() {
			bidsOfYes = append(bidsOfYes, order)
			wg.Done()
		}()
		go func() {
			sort.SliceStable(bidsOfYes, func(i, j int) bool {
				if bidsOfYes[i].Price != bidsOfYes[j].Price {
					return bidsOfYes[i].Price < bidsOfYes[j].Price
				}
				return bidsOfYes[i].Qty < bidsOfYes[j].Qty
			})
			wg.Done()
		}()
		wg.Wait()
	} else {
		bidsOfNo = append(bidsOfNo, order)
		sort.SliceStable(bidsOfNo, func(i, j int) bool {
			if bidsOfNo[i].Price != bidsOfNo[j].Price {
				return bidsOfNo[i].Price < bidsOfNo[j].Price
			}
			return bidsOfNo[i].Qty < bidsOfNo[j].Qty
		})
	}

	var orderBookBids Orderbook
	orderBookBids.BidYes = bidsOfYes
	orderBookBids.BidNo = bidsOfNo
	return h.SendResponse(c, h.Response{
		Code:    fiber.StatusOK,
		Message: "Match not found",
		Data:    orderBookBids,
	})
}

func findMatchWithSet(side string, price float32, userId string, qty int16, eventId string) (int16, error) {
	getPrice := 10 - price
	var getSide string
	if side == "YES" {
		getSide = "NO"
	} else {
		getSide = "YES"
	}
	redisKey := fmt.Sprintf("%s:bids%s:%g", eventId, getSide, getPrice)

	redisclient.Rdb.ZRem()

	data, err := redisclient.Rdb.ZRangeByScore(ctx, redisKey, &redis.ZRangeBy{
		Min:   fmt.Sprintf("%f", float64(qty)),
		Max:   fmt.Sprintf("%f", float64(qty)),
		Count: 1,
	}).Result()
	if err != nil {
		fmt.Println("Err>>", err)
	}
	if len(data) == 0 {
		return 0, nil
	}
	var order Order
	err = h.Unmarshal(data[0], &order)
	if err != nil {
		return 0, err
	}

	fmt.Println(order)
	return 1, nil
}

func MakeOrderWithRedis(c *fiber.Ctx) error {
	var body RequestBody
	err := c.BodyParser(&body)
	if err != nil {
		return h.ErrorHandler(c, err)
	}
	// handle enum
	m := make(map[string]string)
	m["side"] = body.Side.Valid()
	m["purchaseFlag"] = body.PurchaseFlag.Valid()

	if err := h.ValidateEnum(m); len(err) != 0 {
		return h.SendResponse(c, h.Response{
			Code:    fiber.StatusBadRequest,
			Message: "Enum validation failed",
			Data:    err,
		})
	}

	// matching the order
	result, err := findMatchWithSet(string(body.Side), body.Price, body.UserID, body.Qty, body.EventID)
	if err != nil {
		return h.ErrorHandler(c, err)
	}
	if result == 1 {
		return h.SendResponse(c, h.Response{
			Code:    fiber.StatusOK,
			Message: "Match found",
			Data:    nil,
		})
	}

	// if not match then add to the set
	redisKey := fmt.Sprintf("%s:bids%s:%g", body.EventID, body.Side, body.Price)
	uuid := uuid.New()
	var order Order
	order.ID = uuid
	order.UserId = body.UserID
	order.Qty = body.Qty
	order.Price = body.Price

	jsonString, _ := h.Marshal(order)
	_, err = redisclient.Rdb.ZAdd(ctx, redisKey, redis.Z{Member: jsonString, Score: float64(body.Qty)}).Result()
	if err != nil {
		return h.ErrorHandler(c, err)
	}

	return h.SendResponse(c, h.Response{
		Code:    fiber.StatusOK,
		Message: "Not match",
		Data:    order,
	})
}

// // Define the batch size and the number of goroutines
// 	batchSize := 50000
// 	numGoroutines := 10

// 	// Create a channel to coordinate goroutines
// 	ch := make(chan int)

// 	for i := 0; i < numGoroutines; i++ {
// 		go func(id int) {
// 			// Calculate the range for this goroutine
// 			start := id * batchSize
// 			end := (id + 1) * batchSize

// 			// Create a pipeline for this goroutine
// 			pipe := redisclient.Rdb.Pipeline()

// 			for j := start; j < end; j++ {
// 				var order Order
// 				uuid := uuid.New()
// 				order.ID = uuid
// 				jsonData, _ := json.Marshal(order)
// 				// if err != nil {
// 				// 	// Handle the error
// 				// }

// 				// Queue the ZAdd command in the pipeline
// 				pipe.ZAdd(ctx, "bidsOfYes", &redis.Z{Member: jsonData, Score: float64(j)})
// 			}

// 			// Execute the pipeline for this batch
// 			_, err := pipe.Exec(ctx)
// 			if err != nil {
// 				ch <- 0
// 			}

// 			// Notify that the goroutine has finished its work
// 			ch <- id
// 		}(i)
// 	}

// 	// Wait for all goroutines to finish
// 	for i := 0; i < numGoroutines; i++ {
// 		<-ch
// 	}
