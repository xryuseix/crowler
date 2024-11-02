package main

import (
	"context"
	"fmt"
	"math/rand"
	"xryuseix/crawler/app/fetch"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var redisClient = redis.NewClient(&redis.Options{
	// TODO: Change to `redis:6379`
    Addr: "localhost:6379",
})

type Subscriber struct {
	moving *redis.PubSub
	thread *redis.PubSub
}

type Channel struct {
	moving string
	thread string
}
var channel = Channel{
	moving: "moving",
	thread: "thread",
}

func getUrl() string {
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}
	return urls[rand.Intn(len(urls))]
}

func receiveMoving(msg string) {
	fmt.Println("msg", msg)
	html := fetch.FakeFetch(getUrl())
	if err := redisClient.Publish(ctx, channel.moving, rand.Intn(100)).Err(); err != nil {
        panic(err)
    }
	fmt.Println("fetched", html)
}

func receiveThreadFree(msg string) {
	fmt.Println("msg", msg)
	html := fetch.FakeFetch(getUrl())
	if err := redisClient.Publish(ctx, channel.thread, rand.Intn(10)).Err(); err != nil {
        panic(err)
    }
	fmt.Println("fetched", html)
}

func (s *Subscriber) receiveMessage(ctx context.Context, quit chan int) (error) {
	select {
	case msg := <-s.moving.Channel():
		receiveMoving(msg.Payload)
		return nil
	case msg := <-s.thread.Channel():
		receiveThreadFree(msg.Payload)
		return nil
	case <-ctx.Done():
		quit <- 1
		return ctx.Err()
	}
}

func subscribe(ready, quit chan int) {
	subscriber := Subscriber{
		moving: redisClient.Subscribe(ctx, channel.moving),
		thread: redisClient.Subscribe(ctx, channel.thread),
	}
	ready <- 1

	for {
		err := subscriber.receiveMessage(ctx, quit)
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
	if err := redisClient.Publish(ctx, channel.moving, 100).Err(); err != nil {
        panic(err)
    }
	if err := redisClient.Publish(ctx, channel.thread, 3).Err(); err != nil {
        panic(err)
    }

	<-quit
}
