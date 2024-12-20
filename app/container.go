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

func (c *Container) Start() {
	for {
		if !c.running {
			return
		}
		time.Sleep(time.Duration(config.Configs.WaitTime) * time.Second)
		fmt.Printf("[%d] running...\n", c.id)
		q, err := c.DeQueueingURL()
		if err != nil {
			log.Print(err)
			c.running = false
			continue
		}
		if q.Hops > config.Configs.Hops {
			continue
		}

		v, nq, err := c.Fetch(&q)
		if err != nil {
			log.Print(err)
			continue
		}
		for i := range nq {
			nq[i].Hops = q.Hops + 1
		}

		if err := c.InsertFetchResultToDB(v); err != nil {
			log.Print(err)
			continue
		}
		if err := c.QueueingURL(nq); err != nil {
			log.Print(err)
			continue
		}
	}
}

func (c *Container) Stop() {
	c.running = false
	fmt.Printf("[%d] stopped\n", c.id)
}

func (c *Container) Fetch(q *Queue) (Visited, []*Queue, error) {
	url, err := url.Parse(q.URL)
	if err != nil {
		log.Print(err)
		return Visited{}, []*Queue{}, err
	}

	fmt.Printf("[%d] fetching %s\n", c.id, url.String())

	p := fetch.NewParser(url)
	if err := p.GetWebPage(); err != nil {
		log.Print(err)
		return Visited{}, []*Queue{}, err
	}
	if err := p.Parse(); err != nil {
		log.Print(err)
		return Visited{}, []*Queue{}, err
	}
	fm := fetch.NewFileManager(p.CDP.HTML, p.ResourceLinks)
	fm.ReplaceLinks()

	d := fetch.NewDownloader(url, p.CDP.Shot, fm)
	d.DownloadFiles()

	queues := make([]*Queue, 0, len(p.Links))
	for _, link := range p.Links {
		u, err := url.Parse(link)
		if err != nil {
			log.Print(err)
			continue
		}
		queues = append(queues, &Queue{
			URLHash:  lib.Hash(u.String()),
			URL:      u.String(),
			Domain:   u.Host,
			SeedFile: q.SeedFile,
		})
	}

	// TODO: URLからparamを消す
	v := Visited{
		URLHash: lib.Hash(url.String()),
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

	var validq []Queue
	if config.Configs.Duplicate == "same-url" {
		urls := make([]string, len(queues))
		for i, q := range queues {
			urls[i] = q.URLHash
		}
		var v []Visited
		if err := c.db.Where("url_hash IN (?)", urls).Find(&v).Error; err != nil {
			return err
		}
		vmap := make(map[string]bool, len(v))
		for _, vv := range v {
			vmap[vv.URLHash] = true
		}

		for _, q := range queues {
			if _, ok := vmap[q.URLHash]; ok {
				continue
			}
			validq = append(validq, *q)
		}
	} else if config.Configs.Duplicate == "same-domain" {
		queues = c.UniqueDomain(queues)
		hosts := make([]string, len(queues))
		for i, q := range queues {
			u, err := url.Parse(q.URL)
			if err != nil {
				log.Print(err)
				continue
			}
			hosts[i] = u.Host
		}
		hosts = lib.Unique(hosts)

		dupmap := make(map[string]bool)

		var dupq []Queue
		if err := c.db.Where("domain IN (?)", hosts).Find(&dupq).Error; err != nil {
			return err
		}
		for _, q := range dupq {
			dupmap[q.Domain] = true
		}

		var dupv []Visited
		if err := c.db.Where("domain IN (?)", hosts).Find(&dupv).Error; err != nil {
			return err
		}
		for _, v := range dupv {
			dupmap[v.Domain] = true
		}

		for _, q := range queues {
			if _, ok := dupmap[q.Domain]; ok {
				continue
			}
			validq = append(validq, *q)
		}
	} else {
		validq = make([]Queue, len(queues))
		for i, q := range queues {
			validq[i] = *q
		}
	}
	if err := c.db.Create(&validq).Error; err != nil {
		return err
	}
	return nil
}

func (c *Container) UniqueDomain(q []*Queue) []*Queue {
	vsm := make(map[string]struct{})
	r := make([]*Queue, 0)
	for _, qq := range q {
		if _, ok := vsm[qq.Domain]; ok {
			continue
		}
		vsm[qq.Domain] = struct{}{}
		r = append(r, qq)
	}
	return r
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
