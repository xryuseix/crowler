package fetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"xryuseix/crawler/app/lib"

	"github.com/PuerkitoBio/goquery"
)

type InternalUrl struct {
	from string
	to   string
}

type Parser struct {
	url          *url.URL
	HTML         string
	Links        []string
	InternalUrls []InternalUrl
}

func NewParser(url *url.URL) *Parser {
	return &Parser{
		url:          url,
		HTML:         "",
		Links:        make([]string, 0),
		InternalUrls: make([]InternalUrl, 0),
	}
}

func (p *Parser) GetWebPage(url *url.URL) error {
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	p.HTML = string(body)
	return nil
}

func (p *Parser) Parse() error {
	if p.HTML == "" {
		return fmt.Errorf("HTML is empty")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	if err != nil {
		return err
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
	resourcesLinks = lib.Filter(resourcesLinks, notNull)

	for _, link := range resourcesLinks {
		if strings.HasPrefix(link, "data:") {
			continue
		}
		if strings.Contains(link, " ") {
			link = strings.Split(link, " ")[0]
		}
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			p.Links = append(p.Links, link)
			continue
		}

		u, err := url.Parse(link)
		if err != nil {
			log.Fatal(err)
		}
		newUrl := p.url.ResolveReference(u).String()
		p.Links = append(p.Links, newUrl)
		p.InternalUrls = append(p.InternalUrls, InternalUrl{
			from: link,
			to:   newUrl,
		})
	}
	return nil
}

func (p *Parser) Url2filename() (string, string) {
	pathParts := strings.Split(strings.TrimPrefix(p.url.Path, "/"), "/")
	return strings.Join(pathParts[:len(pathParts)-1], "/"), pathParts[len(pathParts)-1]
}

func (p *Parser) ReplaceInternalDomains(html string) string {
	for _, domain := range p.InternalUrls {
		html = strings.ReplaceAll(html, domain.from, domain.to)
	}
	return html
}
