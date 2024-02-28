package db

import (
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

func GetRedisClient() *redis.Client {
	// Redis connection details
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Establish Redis connection
	client := redis.NewClient(&redis.Options{
		Addr:               redisAddr,
		Password:           redisPassword,
		DB:                 0,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolSize:           10,
		MinIdleConns:       5,
		MaxConnAge:         0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
	})
	return client
}
