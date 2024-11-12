package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/cdproto/dom"
	// "github.com/chromedp/cdproto/network"
	cdp "github.com/chromedp/chromedp"
)

type CDPRes struct {
	HTML string
	Shot []byte
}

func GetHTMLandSS(_ctx context.Context, url string) (CDPRes, error) {
	ctx, cancel := cdp.NewContext(_ctx)
	defer cancel()

	var filebyte []byte
	var html string
	if err := cdp.Run(ctx, cdp.Tasks{
		cdp.Navigate(url),
		cdp.Sleep(3 * time.Second),
		cdp.CaptureScreenshot(&filebyte),
		cdp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	}); err != nil {
		return CDPRes{}, err
	}
	return CDPRes{HTML: html, Shot: filebyte}, nil
}

func init() {
	if _, err := os.Stat("./out"); os.IsNotExist(err) {
		os.Mkdir("./out", os.ModePerm)
	}
}

func main() {
	fmt.Println("start")
	url := "https://google.com"
	ctx, cancel := cdp.NewContext(context.Background())
	defer cancel()

	res, err := GetHTMLandSS(ctx, url)
	if err != nil {
		panic(err)
	}

	pngFile, err := os.Create("./out/shot.png")
	defer pngFile.Close()
	if err != nil {
		panic(err)
	}

	pngFile.Write(res.Shot)
	fmt.Println("screen shot tacked!")

	for {
		time.Sleep(1 * time.Second)
		fmt.Println("waiting...")
	}
}
