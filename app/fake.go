package main

import (
	"fmt"
	"math/rand"
	"time"
)

func getUrl() string {
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}
	return urls[rand.Intn(len(urls))]
}

func fakeFetch(url string) string {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	fmt.Println("fake fetched")
	return fmt.Sprintf("<html>%s</html>", url)
}
