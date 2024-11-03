package main

import (
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

func (s *Subscriber) receiveMessage( msg string) {
	fmt.Println("msg", msg)

	f := func() {
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
				URL: link,
			})
		}
		fmt.Println("queues", queues, p.Links)
		s.db.Create(&queues)
		s.db.Create(&Visited{
			URL:      url.String(),
			Domain:   url.Host,
			SavePath: fmt.Sprintf("%s/%s", dir, file),
		})
	}
	go f()
}
