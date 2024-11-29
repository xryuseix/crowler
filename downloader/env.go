package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	ServerUser   string
	ServerIP     string
	IdentityPath string
	RemotePath   string
	LocalPath    string
	MaxWorkers   int
}

var env *Env

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	maxWorkers, err := strconv.Atoi(os.Getenv("MAX_WORKERS"))
	if err != nil {
		maxWorkers = 1
	}

	env = &Env{
		ServerUser:   os.Getenv("SERVER_USER"),
		ServerIP:     os.Getenv("SERVER_IP"),
		IdentityPath: os.Getenv("IDENTITY_PATH"),
		RemotePath:   os.Getenv("REMOTE_PATH"),
		LocalPath:    os.Getenv("LOCAL_PATH"),
		MaxWorkers:   maxWorkers,
	}
}
