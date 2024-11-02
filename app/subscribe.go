package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-redis/redis/v8"
)

type Subscriber struct {
	moving *redis.PubSub
	thread *redis.PubSub
}

type Channel struct {
	moving string
	thread string
}

func (s *Subscriber) receiveMessage(ctx context.Context, quit chan int, thread *Thread) error {
	select {
	case msg := <-s.moving.Channel():
		s.receiveMoving(thread, msg.Payload)
		return nil
	case msg := <-s.thread.Channel():
		s.receiveThreadFree(thread, msg.Payload)
		return nil
	case <-ctx.Done():
		quit <- 1
		return ctx.Err()
	}
}

func (s *Subscriber) receiveMoving(thread *Thread, msg string) {
	fmt.Println("msg", msg)

	if err := thread.Dec(); err != nil {
		fmt.Println(err)
		return
	}

	fetch := func(t *Thread) {
		html := fakeFetch(getUrl())
		if err := redisClient.Publish(ctx, channel.moving, rand.Intn(100)).Err(); err != nil {
			panic(err)
		}
		fmt.Println("fetched", html)
		if err := t.Inc(); err != nil {
			fmt.Println(err)
			return
		}
	}
	go fetch(thread)
}

func (s *Subscriber) receiveThreadFree(thread *Thread, msg string) {
	fmt.Println("msg", msg)

	if err := thread.Dec(); err != nil {
		fmt.Println(err)
		return
	}

	fetch := func(t *Thread) {
		html := fakeFetch(getUrl())
		if err := redisClient.Publish(ctx, channel.thread, rand.Intn(10)).Err(); err != nil {
			panic(err)
		}
		fmt.Println("fetched", html)
		if err := t.Inc(); err != nil {
			fmt.Println(err)
			return
		}
	}
	go fetch(thread)
}
