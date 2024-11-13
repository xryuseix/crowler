package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slices"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type chromedpRes struct {
	HTML       string
	Shot       []byte
	requestURL []string
}

func GetExecutor(ctx context.Context) context.Context {
	c := chromedp.FromContext(ctx)
	return cdp.WithExecutor(ctx, c.Target)
}

func GetHTMLandSS(url string) (chromedpRes, []error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.DisableGPU,
        chromedp.WindowSize(1920, 1080),
    )
	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel2 := chromedp.NewContext(
		allocCtx,
		// chromedp.WithDebugf(log.Printf),
	)
	for _, cancel := range []context.CancelFunc{cancel1, cancel2} {
        defer cancel()
    }

	var requestURL []string
	var errors []error
	var capList = []network.ResourceType{"Document", "Stylesheet", "Image", "Media", "Font", "Script"}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func(ev *fetch.EventRequestPaused) {
				if !slices.Contains(capList, ev.ResourceType) {
					return
				}
				requestURL = append(requestURL, ev.Request.URL)
				r := fetch.ContinueRequest(ev.RequestID)
				if err := r.Do(GetExecutor(ctx)); err != nil {
					errors = append(errors, err)
				}
			}(ev)
		}
	})

	var filebyte []byte
	var html string
	if err := chromedp.Run(ctx, chromedp.Tasks{
		fetch.Enable(),
		chromedp.Navigate(url),
		chromedp.Sleep(3 * time.Second),
		chromedp.CaptureScreenshot(&filebyte),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	}); err != nil {
		errors = append(errors, err)
		return chromedpRes{}, errors
	}
	return chromedpRes{HTML: html, Shot: filebyte, requestURL: requestURL}, nil
}

func init() {
	if _, err := os.Stat("./out"); os.IsNotExist(err) {
		os.Mkdir("./out", os.ModePerm)
	}
}

func main() {
	fmt.Println("start")
	// url := "https://google.com"
	url := "http://localhost:8080"

	res, errors := GetHTMLandSS(url)
	if len(errors) > 0 {
		panic(errors)
	}

	pngFile, err := os.Create("./out/shot.png")
	defer pngFile.Close()
	if err != nil {
		panic(err)
	}

	pngFile.Write(res.Shot)
	fmt.Println("screen shot tacked!")
	fmt.Printf("HTML len: %d\n", len(res.HTML))
	fmt.Printf("requestURL: %s\n", res.requestURL)
}
