package main

import (
	"context"
	"fmt"
	"net/url"

	"xryuseix/crawler/app/fetch"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Subscriber struct {
	moving *redis.PubSub
	thread *redis.PubSub
	db     *gorm.DB
}

type Channel struct {
	moving string
	thread string
}

func (s *Subscriber) receiveMessageMngr(ctx context.Context, quit chan int, thread *Thread) error {
	select {
	case msg := <-s.moving.Channel():
		s.receiveMessage(thread, msg.Payload)
		return nil
	case msg := <-s.thread.Channel():
		s.receiveMessage(thread, msg.Payload)
		return nil
	case <-ctx.Done():
		quit <- 1
		return ctx.Err()
	}
}

func (s *Subscriber) receiveMessage(thread *Thread, msg string) {
	fmt.Println("msg", msg)

	if err := thread.Dec(); err != nil {
		fmt.Println(err)
		return
	}

	f := func(t *Thread) {
		_url := getFakeUrl()
		url, err := url.Parse(_url)
		if err != nil {
			fmt.Println(err)
			return
		}
		html := fakeFetch(url)
		fmt.Println("fetched", html)

		p := fetch.NewParser(url)
		p.HTML = html
		if err := p.Parse(); err != nil {
			fmt.Println(err)
			return
		}
		p.HTML = p.ReplaceInternalDomains(p.HTML)
		dir, file := p.Url2filename()

		queues := make([]Queue, 0, len(p.Links))
		for _, link := range p.Links {
			queues = append(queues, Queue{
				url: link,
			})
		}
		fmt.Println("queues", queues, p.Links)
		s.db.Create(&queues)
		s.db.Create(&Visited{
			url:      url.String(),
			domain:   url.Host,
			savePath: fmt.Sprintf("%s/%s", dir, file),
		})

		if err := t.Inc(); err != nil {
			fmt.Println(err)
			return
		}
	}
	go f(thread)
}
