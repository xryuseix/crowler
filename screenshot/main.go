package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/exp/slices"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
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
	// TODO: with timeout https://github.com/chromedp/chromedp/issues/1009
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-cache", true),
	)
	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel2 := chromedp.NewContext(
		allocCtx,
		chromedp.WithDebugf(log.Printf),
	)
	for _, cancel := range []context.CancelFunc{cancel1, cancel2} {
		defer cancel()
	}

	var requestURL []string
	var errors []error
	var skipList = []network.ResourceType{"Fetch", "XHR"}
	var capList = []network.ResourceType{"Document", "Stylesheet", "Image", "Media", "Font", "Script"}
	ch := make(chan bool)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func(ev *fetch.EventRequestPaused) {
				inCap := slices.Contains(capList, ev.ResourceType)
				inSkip := slices.Contains(skipList, ev.ResourceType)
				if !inCap && !inSkip {
					return
				}
				if inCap {
					requestURL = append(requestURL, ev.Request.URL)
				}
				r := fetch.ContinueRequest(ev.RequestID)
				if err := r.Do(GetExecutor(ctx)); err != nil {
					errors = append(errors, err)
				}
			}(ev)
		case *page.EventLoadEventFired:
			go func() {
				ch <- true
			}()
		}
	})

	if err := chromedp.Run(ctx, chromedp.Tasks{
		fetch.Enable(),
		chromedp.Navigate(url),
	}); err != nil {
		errors = append(errors, err)
		return chromedpRes{}, errors
	}

	<-ch

	var filebyte []byte
	var html string
	if err := chromedp.Run(ctx, chromedp.Tasks{
		// TODO: ScrollIntoView
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
	time.Sleep(10 * time.Second)
	return chromedpRes{HTML: html, Shot: filebyte, requestURL: requestURL}, nil
}

func init() {
	if _, err := os.Stat("./out"); !os.IsNotExist(err) {
		os.RemoveAll("./out")
	}
	if _, err := os.Stat("./out"); os.IsNotExist(err) {
		os.Mkdir("./out", os.ModePerm)
	}
}

func main() {
	fmt.Println("start")
	// url := "https://google.com"
	url := "http://127.0.0.1:5000"

	res, errors := GetHTMLandSS(url)
	if len(errors) > 0 {
		fmt.Println(errors)
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

	out, err := os.Create("./out/index.html")
	defer out.Close()
	if err != nil {
		panic(err)
	}
	out.Write([]byte(res.HTML))
}