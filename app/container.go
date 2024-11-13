package main

import (
	"fmt"
	"net/url"
	"time"

	"xryuseix/crowler/app/fetch"

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
		time.Sleep(2 * time.Second)
		fmt.Printf("[%d] running...\n", c.id)
		q, err := c.DeQueueingURL()
		if err != nil {
			fmt.Println(err)
			continue
		}

		v, nq, err := c.Fetch(q.URL)
		if err != nil {
			fmt.Println(err)
			continue
		}

		c.InsertFetchResultToDB(v)
		c.QueueingURL(nq)
	}
}

func (c *Container) Stop() {
	c.running = false
	fmt.Printf("[%d] stopped\n", c.id)
}

func (c *Container) Fetch(_url string) (Visited, []*Queue, error) {
	url, err := url.Parse(_url)
	if err != nil {
		fmt.Println(err)
		return Visited{}, []*Queue{}, err
	}

	fmt.Printf("[%d] fetching %s\n", c.id, url.String())

	p := fetch.NewParser(url)
	if err := p.GetWebPage(); err != nil {
		fmt.Println(err)
		time.Sleep(time.Second)
		return Visited{}, []*Queue{}, err
	}
	if err := p.Parse(); err != nil {
		fmt.Println(err)
		time.Sleep(time.Second)
		return Visited{}, []*Queue{}, err
	}
	p.CDP.HTML = p.ReplaceInternalDomains(p.CDP.HTML)

	d := fetch.NewDownloader(url, p.ResourceLinks, p.CDP.HTML, p.CDP.Shot)
	d.DownloadFiles()

	queues := make([]*Queue, 0, len(p.Links))
	for _, link := range p.Links {
		queues = append(queues, &Queue{
			URL: link,
		})
	}

	// TODO: URLからparamを消す
	v := Visited{
		URL:     url.String(),
		Domain:  url.Host,
		SaveDir: d.SaveDir,
	}

	return v, queues, nil
}

func (c *Container) QueueingURL(q []*Queue) {
	// INSERT INTO queues (url) VALUES ('https://example.com');
	c.db.Create(&q)
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

	if queue.Id == 0 && queue.URL == "" {
		return Queue{}, fmt.Errorf("queue is empty")
	}
	return queue, nil
}

func (c *Container) InsertFetchResultToDB(v Visited) {
	c.db.Create(&v)
}
