package main

import (
	"fmt"
	"log"
	"net/url"

	"xryuseix/crowler/app/fetch"
	"xryuseix/crowler/app/config"
)

func init() {
	if err := config.LoadConf("../config.yaml"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	_url := "https://example.com"
	url, err := url.Parse(_url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("[INFO] fetching %s\n", url.String())

	p := fetch.NewParser(url)
	if err := p.GetWebPage(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(p.CDP.HTML)

	if err := p.Parse(); err != nil {
		log.Fatal(err)
	}
	fm := fetch.NewFileManager(p.CDP.HTML, p.ResourceLinks)
	fm.ReplaceLinks()

	d := fetch.NewDownloader(url, p.CDP.Shot, fm)
	d.DownloadFiles()
}