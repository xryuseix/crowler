package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Parser struct {
	url             string
	html            string
	links           []string
	externalDomains []string
}

func NewParser(url string) *Parser {
	return &Parser{
		url:             url,
		html:            "",
		links:           make([]string, 0),
		externalDomains: make([]string, 0),
	}
}

func (p *Parser) GetWebPage(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	p.html = string(body)
	return nil
}

func (p *Parser) Parse() {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p.html))
	if err != nil {
		log.Fatal(err)
	}

	var resourcesLinks []string
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		resourcesLinks = append(resourcesLinks, link)
	})
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		srcset, _ := s.Attr("srcset")
		resourcesLinks = append(resourcesLinks, src, srcset)
	})
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		resourcesLinks = append(resourcesLinks, src)
	})

	// Removing any empty links
	notNull := func(v string) bool {
		return v != ""
	}
	resourcesLinks = filter(resourcesLinks, notNull)

	for _, link := range resourcesLinks {
		if strings.HasPrefix(link, "data:") {
			continue
		}
		if strings.Contains(link, " ") {
			link = strings.Split(link, " ")[0]
		}

		u, err := url.Parse(link)
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			p.externalDomains = append(p.externalDomains, fmt.Sprintf("%s://%s", u.Scheme, u.Host))
			p.links = append(p.links, link)
		} else {
			newUrl := u.ResolveReference(u).String()
			p.links = append(p.links, newUrl)
		}
	}
}
