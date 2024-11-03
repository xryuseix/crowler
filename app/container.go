package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Container struct {
	id int
	db *gorm.DB
	running bool
}

func NewContainer(id int, db *gorm.DB) *Container {
	fmt.Printf("[%d] created\n", id)
	return &Container{ 
		id: id,
		db: db,
		running: true,
	}
}

func (c *Container) Start() error {
	for {
		if !c.running {
			return nil
		}
		fmt.Printf("[%d] running...\n", c.id)
		time.Sleep(1 * time.Second)
	}
}

func (c *Container) Stop() {
	c.running = false
	fmt.Printf("[%d] stopped\n", c.id)
}


// func (c *Container) Fetch() error {
// 	_url := getFakeUrl()
// 		url, err := url.Parse(_url)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		html := fakeFetch(url)
// 		fmt.Println("fetched", html)
// }