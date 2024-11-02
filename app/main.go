package main

import (
	"context"
	"fmt"
	"sync"
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

var THREAD_MAX = 8
type Thread struct {
	mu sync.Mutex
	left int
}
func (t *Thread) Inc() error {
	t.mu.Lock()
	if t.left == THREAD_MAX {
		t.mu.Unlock()
		return fmt.Errorf("no thread left")
	}
	t.left++
	t.mu.Unlock()
	return nil
}
func (t *Thread) Dec() error {
	t.mu.Lock()
	if t.left == 0 {
		t.mu.Unlock()
		return fmt.Errorf("no thread left")
	}
	t.left--
	t.mu.Unlock()
	return nil
}

var thread = Thread{
	left: THREAD_MAX,
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
	
	if err := thread.Dec(); err != nil {
		fmt.Println(err)
		return
	}
	
	fetch := func(t *Thread) {
		html := fetch.FakeFetch(getUrl())
		if err := redisClient.Publish(ctx, channel.moving, rand.Intn(100)).Err(); err != nil {
			panic(err)
		}
		fmt.Println("fetched", html)
		if err := t.Inc(); err != nil {
			fmt.Println(err)
			return
		}
	}
	go fetch(&thread)
}

func receiveThreadFree(msg string) {
	fmt.Println("msg", msg)

	if err := thread.Dec(); err != nil {
		fmt.Println(err)
		return
	}

	fetch := func(t *Thread) {
		html := fetch.FakeFetch(getUrl())
		if err := redisClient.Publish(ctx, channel.thread, rand.Intn(10)).Err(); err != nil {
			panic(err)
		}
		fmt.Println("fetched", html)
		if err := t.Inc(); err != nil {
			fmt.Println(err)
			return
		}
	}
	go fetch(&thread)
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
	for i := 0; i < THREAD_MAX; i++ {
		if err := redisClient.Publish(ctx, channel.thread, i).Err(); err != nil {
			panic(err)
		}
	}

	<-quit
}
