package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xryuseix/crowler/app/config"
	"xryuseix/crowler/app/lib"
)

type Queue struct {
	Id       int    `gorm:"primaryKey;autoIncrement;not null"`
	URLHash  string `gorm:"unique;not null"`
	URL      string `gorm:"not null"`
	Domain   string `gorm:"not null"`
	Hops     int    `gorm:"default:0"`
	SeedFile string `gorm:""`
}

type Visited struct {
	URLHash  string `gorm:"primaryKey;unique;not null"`
	URL      string `gorm:"not null"`
	Domain   string `gorm:"not null"`
	SaveDir  string `gorm:"not null"`
	Hops     int    `gorm:"not null"`
	SeedFile string `gorm:""`
}

func CreateDB() (*gorm.DB, error) {
	e := config.Envs
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo", e.DBHost, e.User, e.Password, e.DBName, e.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		CreateBatchSize: 1000,
		Logger:          logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

var Models = []interface{}{
	&Queue{},
	&Visited{},
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(Models...)
}

func BuildDB() (*gorm.DB, error) {
	db, err := CreateDB()
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

func InsertSeed(db *gorm.DB) {
	if len(config.Configs.SeedFiles) == 0 {
		return
	}

	c := NewContainer(-1, db)

	for _, f := range config.Configs.SeedFiles {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		l := lib.Unique(strings.Split(string(b), "\n"))
		l = lib.Filter(l, func(s string) bool { return s != "" })

		var urls []*url.URL
		for _, u := range l {
			u, err := url.Parse(u)
			if err != nil {
				log.Print(err)
				continue
			}
			urls = append(urls, u)
		}

		var q []*Queue = make([]*Queue, len(urls))
		for i, u := range urls {
			q[i] = &Queue{
				URLHash:  lib.Hash(u.String()),
				URL:      u.String(),
				Domain:   u.Host,
				SeedFile: f,
			}
		}
		c.QueueingURL(q)
	}
}

func InsertRandomSeed(db *gorm.DB) {
	if config.Configs.RandomSeed == false {
		return
	}
	n := config.Configs.ThreadMax * 2
	var q []*Queue = make([]*Queue, 0, n)
	for i := 0; i < n; i++ {
		u, err := url.Parse(fmt.Sprintf("https://www.google.com/search?q=%d", rand.Intn(100)))
		if err != nil {
			log.Fatal(err)
		}
		q = append(q, &Queue{
			URLHash: lib.Hash(u.String()),
			URL:     u.String(),
			Domain:  u.Host,
		})
	}
	db.Create(&q)
}
