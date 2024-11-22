package fetch

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"xryuseix/crowler/app/chromedp"
	"xryuseix/crowler/app/lib"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

type ExternalUrl struct {
	from string
	to   string
}

type ResourceLink struct {
	from string
	to   string
}

type Parser struct {
	url *url.URL
	CDP *chromedp.ChromeDP
	// aタグで移動することができるリンク(絶対パス)
	Links []string
	// 画像やスクリプトなどのリソースリンク
	ResourceLinks []ResourceLink
	// リソースリンクのうち、内部リンクを除いたもの
	// 外部リンクは内部リンクに変換する
	ExternalUrls []ExternalUrl
	// 文字数が長いリソースパスを短縮したものを格納するディレクトリ
	TmpDir string
}

func NewParser(url *url.URL) *Parser {
	return &Parser{
		url:           url,
		CDP:           chromedp.NewChromeDP(url),
		Links:         make([]string, 0),
		ResourceLinks: make([]ResourceLink, 0),
		ExternalUrls:  make([]ExternalUrl, 0),
		TmpDir:        fmt.Sprintf("%s", uuid.New().String()),
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
	anchorLinks = lib.ToAbsoluteLink(p.url, anchorLinks)

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
	p.Links = append(p.Links, anchorLinks...)
	p.ResourceLinks = p.ReplaceLongLink(rlinks)
	p.ExternalUrls = append(p.ExternalUrls, reurl...)

	p.Links = lib.Unique(p.Links)
	p.ResourceLinks = lib.Unique(p.ResourceLinks)
	p.ExternalUrls = lib.Unique(p.ExternalUrls)

	return nil
}

func (p *Parser) ReplaceUrls(html string) string {
	for _, domain := range p.ExternalUrls {
		html = strings.ReplaceAll(html, domain.from, domain.to)
	}
	for _, link := range p.ResourceLinks {
		html = strings.ReplaceAll(html, link.from, link.to)
	}
	return html
}

func (p *Parser) ReplaceLongLink(links []string) []ResourceLink {
	r := []ResourceLink{}
	for _, link := range links {
		if len(link) > 255 {
			r = append(r, ResourceLink{
				from: link,
				to:   fmt.Sprintf("/%s/%s", p.TmpDir, uuid.New().String()),
			})
			continue
		}
		r = append(r, ResourceLink{
			from: link,
			to:   link,
		})
	}
	return r
}
