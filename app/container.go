package main

import (
	"fmt"
	"time"
	"math/rand"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Container struct {
	id      int
	db      *gorm.DB
	running bool
}

func NewContainer(id int, db *gorm.DB) *Container {
	fmt.Printf("[%d] created\n", id)
	return &Container{
		id:      id,
		db:      db,
		running: true,
	}
}

func (c *Container) Start() error {
	for {
		if !c.running {
			return nil
		}
		fmt.Printf("[%d] running...\n", c.id)
		time.Sleep(2 * time.Second)
		c.QueueingURL()
		c.QueueingURL()
		c.DeQueueingURL()
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

func (c *Container) QueueingURL() error {
	// INSERT INTO queues (url) VALUES ('https://example.com');
	queue := Queue{
		URL: fmt.Sprintf("https://example.com/%d/%d", c.id, rand.Intn(10000)),
	}
	c.db.Create(&queue)
	return nil
}

func (c *Container) DeQueueingURL() (Queue, error) {
	// DELETE FROM queues WHERE id = (SELECT id FROM queues ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING *;
	queue := Queue{}
	c.db.Clauses(
		clause.Returning{},
	).Where("id = (?)",
		c.db.Table("queues").Select("id").Order("id").Limit(1).Clauses(
			clause.Locking{
				Strength: clause.LockingStrengthUpdate,
				Options:  clause.LockingOptionsSkipLocked,
			},
		),
	).Delete(&queue)

	fmt.Printf("result: %#+v\n", queue)
	return queue, nil
}

func (c *Container) InsertFetchResultToDB() error {
	return nil
}
