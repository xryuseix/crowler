package fetch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Downloader struct {
	base   *url.URL
	links   []string
	html    string
	SaveDir string
}

func NewDownloader(url *url.URL, html string, resourcesLinks []string) *Downloader {
	d := Downloader{}
	saveDir := d.url2dirname(url)
	return &Downloader{
		base:    url,
		links:   resourcesLinks,
		html:    html,
		SaveDir: d.removeTrailingSlash(saveDir),
	}
}

func (d *Downloader) url2dirname(_url *url.URL) string {
	u := _url.Host
	if _url.Path != "" {
		u += url.QueryEscape(_url.Path)
	}
	MAX_PATH_LEN := 32
	if len(u) >= MAX_PATH_LEN {
		u = u[:MAX_PATH_LEN]
	}
	return u
}

func (d *Downloader) DownloadFiles() {
	downloadDir := filepath.Join("out", d.SaveDir, "contents")
	if _, err := os.Stat(d.SaveDir); os.IsNotExist(err) {
		os.MkdirAll(downloadDir, os.ModePerm)
	}
	d.SaveHTML(downloadDir)

	for _, link := range d.links {
		u, err := url.Parse(link)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !(u.Scheme == "http" || u.Scheme == "https") {
			link = d.base.ResolveReference(u).String()
		}
		filePath := filepath.Join(downloadDir, u.RequestURI())
		filePath = filepath.Clean(filePath)

		err = d.download(filePath, link)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (d *Downloader) removeTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return s[:len(s)-1]
	}
	return s
}

func (d *Downloader) download(filePath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Downloading %s -> %s\n", url, filePath)
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (d *Downloader) SaveHTML(downloadDir string) {
	htmlPath := filepath.Join(downloadDir, "index.html")
	out, err := os.Create(htmlPath)
	if err != nil {
		fmt.Println(err)
	}
	out.Write([]byte(d.html))
}
