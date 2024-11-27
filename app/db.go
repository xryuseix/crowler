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
	Id     int    `gorm:"primaryKey;autoIncrement;not null"`
	URL    string `gorm:"unique;not null"`
	Domain string `gorm:"not null"`
	Hops   int    `gorm:"default:0"`
}

type Visited struct {
	URL     string `gorm:"primaryKey;unique;not null"`
	Domain  string `gorm:"not null"`
	SaveDir string `gorm:"not null"`
	Hops    int    `gorm:"not null"`
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
	if config.Configs.SeedFile == "" {
		return
	}
	b, err := os.ReadFile(config.Configs.SeedFile)
	if err != nil {
		log.Fatal(err)
	}
	lines := lib.Unique(strings.Split(string(b), "\n"))
	lines = lib.Filter(lines, func(s string) bool { return s != "" })

	var urls []*url.URL
	var dupMap = make(map[string]bool)
	dup := config.Configs.Duplicate
	for _, s := range lines {
		s = SliceLongerStr(s)
		u, err := url.Parse(s)
		if err != nil {
			log.Print(err)
			continue
		}
		if dup == "same-url" {
			if _, ok := dupMap[u.String()]; ok {
				continue
			}
			dupMap[u.String()] = true
			urls = append(urls, u)
		} else if dup == "same-domain" {
			if _, ok := dupMap[u.Host]; ok {
				continue
			}
			dupMap[u.Host] = true
			urls = append(urls, u)
		} else {
			urls = append(urls, u)
		}
	}

	var q []*Queue = make([]*Queue, len(urls))
	for i, u := range urls {
		q[i] = &Queue{
			URL:    u.String(),
			Domain: u.Host,
		}
	}
	db.Create(&q)
}

func InsertRandomSeed(db *gorm.DB) {
	n := config.Configs.ThreadMax * 2
	var q []*Queue = make([]*Queue, 0, n)
	for i := 0; i < n; i++ {
		u, err := url.Parse(fmt.Sprintf("https://www.google.com/search?q=%d", rand.Intn(100)))
		if err != nil {
			log.Fatal(err)
		}
		q = append(q, &Queue{
			URL:    u.String(),
			Domain: u.Host,
		})
	}
	db.Create(&q)
}

func SliceLongerStr(s string) string {
	// NOTE: https://stackoverflow.com/questions/70123567/index-row-size-2712-exceeds-btree-version-4-maximum-2704-for-index-while-doing
	if len(s) > 1000 {
		log.Printf("URL is too long: %s", s[0:min(100, len(s))])
		s = s[0:min(1000, len(s))]
	}
	return s
}