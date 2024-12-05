package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"xryuseix/crowler/app/config"
)

type ContainerMngr struct {
	mu         sync.Mutex
	containers []*Container
}

func init() {
	if err := config.LoadConf("config.yaml"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	db, err := BuildDB()
	if err != nil {
		log.Fatal(err)
	}
	InsertSeed(db)
	InsertRandomSeed(db)

	var wg sync.WaitGroup
	wgDone := make(chan bool)
	go func() {
		wg.Wait()
		wgDone <- true
	}()
	defer close(wgDone)

	cm := &ContainerMngr{
		containers: make([]*Container, 0, config.Configs.ThreadMax),
	}

	for i := 0; i < config.Configs.ThreadMax; i++ {
		wg.Add(1)
		go func(i int) {
			c := NewContainer(i, db)
			cm.mu.Lock()
			cm.containers = append(cm.containers, c)
			cm.mu.Unlock()
			c.Start()
			wg.Done()
		}(i)
	}

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	for {
		select {
		case <-quit:
			cm.mu.Lock()
			for _, c := range cm.containers {
				fmt.Printf("[%d] stopping...\n", c.id)
				c.Stop()
			}
			cm.mu.Unlock()
		case <-wgDone:
			fmt.Println("INFO: all containers stopped")
			return
		}
	}
}
