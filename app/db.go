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
	var q []*Queue = make([]*Queue, 0)
	b, err := os.ReadFile(config.Configs.SeedFile)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		u, err := url.Parse(line)
		if err != nil {
			log.Print(err)
		}
		q = append(q, &Queue{
			URL:    u.String(),
			Domain: u.Host,
		})
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
