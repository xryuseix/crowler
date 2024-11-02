package main

import (
	"time"
	"log"
)

func main() {
	for {
		log.Println("Mover is running...")
		time.Sleep(30 * time.Second)
	}
}