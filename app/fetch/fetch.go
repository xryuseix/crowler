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

type ExternalUrl struct {
	from string
	to   string
}

type Parser struct {
	url  *url.URL
	HTML string
	// aタグで移動することができるリンク
	Links []string
	// 画像やスクリプトなどのリソースリンク
	ResourceLinks []string
	// リソースリンクのうち、内部リンクを除いたもの
	// 外部リンクは内部リンクに変換する
	ExternalUrls []ExternalUrl
}

func NewParser(url *url.URL) *Parser {
	return &Parser{
		url:           url,
		HTML:          "",
		Links:         make([]string, 0),
		ResourceLinks: make([]string, 0),
		ExternalUrls:  make([]ExternalUrl, 0),
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

	var anchorLinks []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		anchorLinks = append(anchorLinks, link)
	})

	// Removing invalid links
	notNull := func(v string) bool {
		return v != ""
	}
	notDataSchema := func(v string) bool {
		return !strings.HasPrefix(v, "data:")
	}
	resourcesLinks = lib.Filter(lib.Filter(resourcesLinks, notNull), notDataSchema)
	resourcesLinks = lib.SplitBySpace(resourcesLinks)
	anchorLinks = lib.Filter(lib.Filter(anchorLinks, notNull), notDataSchema)
	anchorLinks = lib.SplitBySpace(anchorLinks)

	f := func(links []string) ([]string, []ExternalUrl) {
		l := []string{}
		eu := []ExternalUrl{}
		for _, link := range links {
			l = append(l, link)
			if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
				u, err := url.Parse(link)
				if err != nil {
					log.Fatal(err)
					continue
				}
				eu = append(eu, ExternalUrl{
					from: link,
					to:   fmt.Sprintf(".%s", u.RequestURI()),
				})
			}
		}
		return l, eu
	}

	rlinks, reurl := f(resourcesLinks)
	p.ResourceLinks = append(p.ResourceLinks, rlinks...)
	p.ExternalUrls = append(p.ExternalUrls, reurl...)

	p.Links = append(p.Links, anchorLinks...)

	p.Links = lib.Unique(p.Links)
	p.ResourceLinks = lib.Unique(p.ResourceLinks)
	p.ExternalUrls = lib.Unique(p.ExternalUrls)

	return nil
}

func (p *Parser) ReplaceInternalDomains(html string) string {
	for _, domain := range p.ExternalUrls {
		html = strings.ReplaceAll(html, domain.from, domain.to)
	}
	return html
}
