package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"xryuseix/crowler/app/config"
	"xryuseix/crowler/app/fetch"
	"xryuseix/crowler/app/lib"

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
		time.Sleep(time.Duration(config.Configs.WaitTime) * time.Second)
		fmt.Printf("[%d] running...\n", c.id)
		q, err := c.DeQueueingURL()
		if err != nil {
			log.Fatal(err)
			continue
		}

		v, nq, err := c.Fetch(q.URL)
		if err != nil {
			log.Fatal(err)
			continue
		}

		if err := c.InsertFetchResultToDB(v); err != nil {
			log.Fatal(err)
			continue
		}
		if err := c.QueueingURL(nq); err != nil {
			log.Fatal(err)
			continue
		}
	}
}

func (c *Container) Stop() {
	c.running = false
	fmt.Printf("[%d] stopped\n", c.id)
}

func (c *Container) Fetch(_url string) (Visited, []*Queue, error) {
	url, err := url.Parse(_url)
	if err != nil {
		log.Fatal(err)
		return Visited{}, []*Queue{}, err
	}

	fmt.Printf("[%d] fetching %s\n", c.id, url.String())

	p := fetch.NewParser(url)
	if err := p.GetWebPage(); err != nil {
		log.Fatal(err)
		time.Sleep(time.Second)
		return Visited{}, []*Queue{}, err
	}
	if err := p.Parse(); err != nil {
		log.Fatal(err)
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

func (c *Container) QueueingURL(queues []*Queue) error {
	// INSERT INTO queues (url) VALUES ('https://example.com');
	// WHERE {duplicate config}
	if len(queues) == 0 {
		return nil
	}

	queues = lib.Unique(queues)

	var dupq []Queue
	if config.Configs.Duplicate == "same-url" {
		urls := make([]string, len(queues))
		for i, q := range queues {
			urls[i] = q.URL
		}
		if err := c.db.Where("url IN (?)", urls).Find(&dupq).Error; err != nil {
			return err
		}
	} else if config.Configs.Duplicate == "same-domain" {
		urls := make([]string, len(queues))
		for i, q := range queues {
			u, err := url.Parse(q.URL)
			if err != nil {
				log.Fatal(err)
				continue
			}
			urls[i] = u.Host
		}
		if err := c.db.Where("domain IN (?)", urls).Find(&dupq).Error; err != nil {
			return err
		}
	} else {
		dupq = []Queue{}
	}

	dupu := make(map[string]bool, len(dupq))
	for _, queue := range dupq {
		dupu[queue.URL] = true
	}

	var validq []Queue
	for _, q := range queues {
		if _, ok := dupu[q.URL]; ok {
			continue
		}
		validq = append(validq, *q)
	}

	if err := c.db.Create(&validq).Error; err != nil {
		return err
	}
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

	if queue.Id == 0 && queue.URL == "" {
		return Queue{}, fmt.Errorf("queue is empty")
	}
	return queue, nil
}

func (c *Container) InsertFetchResultToDB(v Visited) error {
	if err := c.db.Create(&v).Error; err != nil {
		return err
	}
	return nil
}
