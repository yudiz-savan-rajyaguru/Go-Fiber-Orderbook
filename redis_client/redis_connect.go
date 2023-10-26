package redisclient

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/opinion-trading/config"
)

var ctx = context.Background()
var Rdb *redis.Client

func Connect() {
	// used for the redis cloud url
	// parsedURL, err := url.Parse(config.ConfigEnv.REDIS_URL)
	// if err != nil {
	// 	panic("Failed to parse Redis URL")
	// }
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.ConfigEnv.REDIS_URL,
		Password: "",
		DB:       0,
	})

	_, err1 := rdb.Ping(ctx).Result()
	if err1 != nil {
		panic("Redis not connect")
	}
	Rdb = rdb
	log.Print("Redis connect success")
}

func FindLength(key string) (string, int64) {
	res, err := Rdb.LLen(ctx, key).Result()
	if err != nil {
		return err.Error(), -1
	}
	return "", res
}
