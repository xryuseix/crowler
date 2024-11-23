package fetch

import (
	"fmt"
	"net/url"
	"strings"

	"xryuseix/crowler/app/chromedp"
	"xryuseix/crowler/app/lib"

	"github.com/PuerkitoBio/goquery"
)

type ResourceLink struct {
	Original string
	Absolute string
}

type Parser struct {
	url *url.URL
	CDP *chromedp.ChromeDP
	// aタグで移動することができるリンク(絶対パス)
	Links []string
	// 画像やスクリプトなどのリソースリンク
	ResourceLinks []ResourceLink
}

func NewParser(url *url.URL) *Parser {
	return &Parser{
		url:           url,
		CDP:           chromedp.NewChromeDP(url),
		Links:         make([]string, 0),
		ResourceLinks: make([]ResourceLink, 0),
	}
}

func (p *Parser) GetWebPage() error {
	if err := p.CDP.GetHTMLAndSS(); err != nil {
		return err
	}
	return nil
}

func (p *Parser) Parse() error {
	if p.CDP.HTML == "" {
		return fmt.Errorf("HTML is empty")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p.CDP.HTML))
	if err != nil {
		return err
	}

	var anchorLinks []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		anchorLinks = append(anchorLinks, link)
	})

	// Removing invalid links
	notNull := func(v string) bool {
		return v != ""
	}
	inValidSchema := func(v string) bool {
		invalid := []string{"data:", "javascript:", "mailto:", "about:"}
		for _, s := range invalid {
			if strings.HasPrefix(v, s) {
				return false
			}
		}
		return true
	}
	resourcesLinks := lib.Filter(lib.Filter(p.CDP.RequestURL, notNull), inValidSchema)
	resourcesLinks = lib.SplitBySpace(resourcesLinks)

	anchorLinks = lib.Filter(lib.Filter(anchorLinks, notNull), inValidSchema)
	anchorLinks = lib.SplitBySpace(anchorLinks)
	anchorLinks = lib.ToAbsoluteLinks(p.url, anchorLinks)

	p.Links = lib.Unique(anchorLinks)
	p.ResourceLinks = make([]ResourceLink, 0, len(resourcesLinks))
	for _, link := range resourcesLinks {
		p.ResourceLinks = append(p.ResourceLinks, ResourceLink{
			Original: link,
			Absolute: lib.ToAbsoluteLink(p.url, link),
		})
	}
	return nil
}
