package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Queue struct {
	Id int    `gorm:"primaryKey;autoIncrement;not null"`
	URL string `gorm:"unique;not null"`
}

type Visited struct {
	URL      string `gorm:"primaryKey;unique;not null"`
	Domain   string `gorm:"not null"`
	SavePath string `gorm:"not null"`
}

func CreateDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo", Envs.DBHost, Envs.User, Envs.Password, Envs.DBName, Envs.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		CreateBatchSize: 1000,
		Logger: logger.Default.LogMode(logger.Info),
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
