package services

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func chache() {
	client := redis.NewClient(
		&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			Protocol: 2,
		},
	)
	fmt.Printf("%v", client.Options().Addr)
}
