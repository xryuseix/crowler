package main

import (
	"fmt"
	"log"
)

func main() {
	url := "https://www.google.com"
	p := NewParser(url)
	err := p.GetWebPage(url)
	if err != nil {
		log.Fatal(err)
	}
	p.Parse()
	fmt.Println(p.links, p.externalDomains)
}
