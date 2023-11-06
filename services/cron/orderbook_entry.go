package cron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/opinion-trading/database"
	redisClient "github.com/opinion-trading/redis_client"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

var ctx = context.Background()

func ProcessOrderbookRedisData() error {
	batchSize := 4
	redisKey := "orderbook"
	backUpKey := "orderbookBackUp"

	list, err := redisClient.Rdb.LRange(ctx, redisKey, 0, int64(batchSize-1)).Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Println("Redis Err>>", err)
		}
		return err
	}

	var orderBook []OrderBookModel
	for _, val := range list {
		var record OrderBookModel
		if err := json.Unmarshal([]byte(val), &record); err != nil {
			return err
		}
		orderBook = append(orderBook, record)
	}

	if err := database.DB.Model(&OrderBookModel{}).Create(&orderBook).Error; err != nil {
		if err.Error() != "empty slice found" {
			// when data is not inserted in db for some issue
			_, err := redisClient.Rdb.RPush(ctx, backUpKey, list).Result()
			if err != nil {
				return errors.New("OrderbookBackUp redis insert fail")
			}
		}
		return err
	}
	// Remove the processed items from the Redis list
	if err := redisClient.Rdb.LTrim(ctx, redisKey, int64(len(list)), -1).Err(); err != nil {
		return err
	}

	return nil
}

func ImplementCron() {
	c := cron.New()

	c.AddFunc("@every 10m", func() {
		err := ProcessOrderbookRedisData()
		if err != nil {
			fmt.Printf("Cron Error: %v\n", err)
		} else {
			fmt.Println("Data transferred from Redis to DB successfully.")
		}
	})

	// Start the cron scheduler
	c.Start()
}
