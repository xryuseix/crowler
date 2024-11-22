package fetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"xryuseix/crowler/app/config"
)

type Downloader struct {
	base    *url.URL
	links   []ResourceLink
	html    string
	shot    []byte
	SaveDir string
}

func NewDownloader(url *url.URL, resourcesLinks []ResourceLink, html string, shot []byte) *Downloader {
	d := Downloader{}
	saveDir := d.url2dirname(url)
	return &Downloader{
		base:    url,
		links:   resourcesLinks,
		html:    html,
		shot:    shot,
		SaveDir: d.rmTrailingSlash(saveDir),
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

func (d *Downloader) DownloadFiles() error {
	fc := config.Configs.FetchContents
	if !fc.Html && !fc.CssJsOther && !fc.ScreenShot {
		return nil
	}

	downloadDir := filepath.Join("out", d.SaveDir)
	contentDir := filepath.Join(downloadDir, "contents")
	if fc.CssJsOther {
		if _, err := os.Stat(d.SaveDir); os.IsNotExist(err) {
			os.MkdirAll(contentDir, os.ModePerm)
		}
	} else if fc.Html || fc.ScreenShot {
		if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
			os.MkdirAll(downloadDir, os.ModePerm)
		}
	}

	if fc.Html {
		if err := d.SaveHTML(downloadDir); err != nil {
			return err
		}
	}

	if fc.ScreenShot {
		if err := d.SaveSS(downloadDir); err != nil {
			return err
		}
	}

	if !fc.CssJsOther {
		return nil
	}

	for _, link := range d.links {
		u, err := url.Parse(link.from)
		if err != nil {
			log.Fatal(err)
			continue
		}

		var filePath string
		if !(u.Scheme == "http" || u.Scheme == "https") {
			link.from = d.base.ResolveReference(u).String()
		}
		if link.from != link.to {
			filePath = filepath.Join(contentDir, link.to)
		} else {
			filePath = filepath.Join(contentDir, u.RequestURI())
		}
		filePath = filepath.Clean(filePath)

		err = d.download(filePath, link.from)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (d *Downloader) rmTrailingSlash(s string) string {
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

	if len(filePath) > 255 {
		return fmt.Errorf("File path is too long: %s", filePath)
	}

	fmt.Printf("Downloading %s -> %s\n", url, filePath)
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	out, err := os.Create(filePath)
	defer out.Close()
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (d *Downloader) SaveHTML(downloadDir string) error {
	htmlPath := filepath.Join(downloadDir, "index.html")
	out, err := os.Create(htmlPath)
	defer out.Close()
	if err != nil {
		return err
	}
	out.Write([]byte(d.html))
	return nil
}

func (d *Downloader) SaveSS(downloadDir string) error {
	htmlPath := filepath.Join(downloadDir, "screenshot.png")
	out, err := os.Create(htmlPath)
	defer out.Close()
	if err != nil {
		return err
	}
	out.Write(d.shot)
	return nil
}
