package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
)

type ContainerMngr struct {
	mu sync.Mutex
	containers []*Container
}

func init() {
	if err := loadConf("config.yaml"); err != nil {
		panic(err)
	}
}

func main() {
	db, err := BuildDB()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	cm := &ContainerMngr{
		containers: make([]*Container, 0, Configs.ThreadMax),
	}
	
	for i := 0; i < Configs.ThreadMax; i++ {
		wg.Add(1)
		go func(){
			c := NewContainer(i, db)
			cm.mu.Lock()
			cm.containers = append(cm.containers, c)
			cm.mu.Unlock()
			c.Start()
		}()
	}
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	for _, c := range cm.containers {
		go func(c *Container) {
			fmt.Printf("[%d] stopping...\n", c.id)
			defer wg.Done()
			c.Stop()
		}(c)
	}

	wg.Wait()
}
