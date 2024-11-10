package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"
)

func getFakeUrl() string {
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}
	return urls[rand.Intn(len(urls))]
}

func fakeFetch(url *url.URL) string {
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	var links int
	// link=0: rand(10) result is 0-4
	// link=1: rand(10) result is 5-7
	// link=2: rand(10) result is 8-9
	// link=3: rand(10) result is 10
	// 期待値: (1*3+2*2+3*1)/10 = 1
	n := rand.Intn(10)
	if n < 5 {
		links = 0
	} else if n < 8 {
		links = 1
	} else if n < 10 {
		links = 2
	} else {
		links = 3
	}

	htmlLink := ""
	for i := 0; i < links; i++ {
		p := rand.Intn(1000)
		htmlLink += fmt.Sprintf("<a href=\"%s?page=%d\">Link</a>", url.String(), p)
	}

	return fmt.Sprintf("<html><body>%s</body></html>", htmlLink)
}
