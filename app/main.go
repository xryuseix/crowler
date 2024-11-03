package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	// TODO: Change to `redis:6379`
	Addr: "redis:6379",
	// Addr: "localhost:6379",
})
var channel = Channel{
	moving: "moving",
	thread: "thread",
}

func subscribe(ready, quit chan int, db *gorm.DB) {
	subscriber := Subscriber{
		moving: rdb.Subscribe(ctx, channel.moving),
		thread: rdb.Subscribe(ctx, channel.thread),
		db:     db,
	}
	ready <- 1

	thread := Thread{
		left: Configs.ThreadMax,
	}

	for {
		err := subscriber.receiveMessageMngr(ctx, quit, &thread)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func init() {
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	if err := loadConf("config.yaml"); err != nil {
		panic(err)
	}
}

func main() {
	db, err := BuildDB()
	if err != nil {
		panic(err)
	}

	quit := make(chan int)
	ready := make(chan int)
	go subscribe(ready, quit, db)

	<-ready
	for i := 0; i < Configs.ThreadMax; i++ {
		if err := rdb.Publish(ctx, channel.thread, i).Err(); err != nil {
			panic(err)
		}
	}

	<-quit
}
