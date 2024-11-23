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
	shot    []byte
	fm      *FileManager
	SaveDir string
}

func NewDownloader(url *url.URL, shot []byte, fm *FileManager) *Downloader {
	d := Downloader{}
	saveDir := d.url2dirname(url)
	return &Downloader{
		shot:    shot,
		fm:      fm,
		SaveDir: saveDir,
	}
}

func (d *Downloader) url2dirname(_url *url.URL) string {
	u := _url.Host
	if _url.Path != "" {
		u += url.QueryEscape(_url.Path)
	}
	MAX_PATH_LEN := 255
	if len(u) >= MAX_PATH_LEN {
		u = u[:MAX_PATH_LEN]
	}
	if strings.HasSuffix(u, "/") {
		return u[:len(u)-1]
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

	for _, link := range d.fm.links {
		var filePath string
		if uuid, ok := d.fm.Table[link.Absolute]; ok {
			filePath = filepath.Join(contentDir, uuid)
		} else {
			continue
		}
		filePath = filepath.Clean(filePath)
		u, err := url.Parse(link.Absolute)
		if err != nil {
			log.Print(err)
			continue
		}
		if err := d.download(filePath, u); err != nil {
			log.Print(err)
			continue
		}
	}
	return nil
}

func (d *Downloader) download(filePath string, url *url.URL) error {
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if len(filePath) > 255 {
		return fmt.Errorf("error: File path is too long: %s", filePath)
	}

	u := url.String()
	fmt.Printf("Downloading %s... -> %s...\n", u[:min(len(u), 64)], filePath[:min(len(filePath), 64)])
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
	oldHtmlPath := filepath.Join(downloadDir, "index.old.html")
	write := func(path string, html string) error {
		out, err := os.Create(path)
		defer out.Close()
		if err != nil {
			return err
		}
		out.Write([]byte(html))
		return nil
	}
	if err := write(htmlPath, d.fm.HTML); err != nil {
		return err
	}
	if err := write(oldHtmlPath, d.fm.OldHtml); err != nil {
		return err
	}
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
