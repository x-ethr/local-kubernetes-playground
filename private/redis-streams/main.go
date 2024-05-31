package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const stream = "demo-stream"
const consumerGroupName = "demo-group"

func main() {

	fmt.Println("redis streams consumer application started")

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("failed to connect", err)
	}

	client.XGroupCreateMkStream(ctx, stream, consumerGroupName, "$")

	for {
		result, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Streams: []string{stream, ">"},
			Group:   consumerGroupName,
			Block:   1 * time.Second,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			log.Fatal("xreadgroup error", err)
		}

		for _, s := range result {
			for _, message := range s.Messages {
				fmt.Println("got data from stream -", message.Values)

				client.XAck(ctx, stream, consumerGroupName, message.ID).Err()
				if err != nil {
					log.Fatal("xack failed for message ", message.ID, err)
				}

				fmt.Println("acknowledged message", message.ID)
			}
		}
	}

}
