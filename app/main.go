package main

import (
	"context"
	"fmt"

	// "xryuseix/crawler/app/fetch"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient = redis.NewClient(&redis.Options{
	// TODO: Change to `redis:6379`
	Addr: "localhost:6379",
})
var channel = Channel{
	moving: "moving",
	thread: "thread",
}

func subscribe(ready, quit chan int) {
	subscriber := Subscriber{
		moving: redisClient.Subscribe(ctx, channel.moving),
		thread: redisClient.Subscribe(ctx, channel.thread),
	}
	ready <- 1

	thread := Thread{
		left: THREAD_MAX,
	}

	for {
		err := subscriber.receiveMessage(ctx, quit, &thread)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func init() {
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	// load config
}

func main() {
	quit := make(chan int)
	ready := make(chan int)
	go subscribe(ready, quit)

	<-ready
	for i := 0; i < THREAD_MAX; i++ {
		if err := redisClient.Publish(ctx, channel.thread, i).Err(); err != nil {
			panic(err)
		}
	}

	<-quit
}
