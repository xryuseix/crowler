package chromedp

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"golang.org/x/exp/slices"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type ChromeDP struct {
	url        *url.URL
	HTML       string
	Shot       []byte
	RequestURL []string
}

func NewChromeDP(url *url.URL) *ChromeDP {
	return &ChromeDP{
		url: url,
	}
}

func (c *ChromeDP) getExecutor(_ctx context.Context) context.Context {
	ctx := chromedp.FromContext(_ctx)
	return cdp.WithExecutor(_ctx, ctx.Target)
}

func (c *ChromeDP) GetHTMLAndSS() error {
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

	var errors []error
	var capList = []network.ResourceType{"Document", "Stylesheet", "Image", "Media", "Font", "Script"}
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func(ev *fetch.EventRequestPaused) {
				if !slices.Contains(capList, ev.ResourceType) {
					return
				}
				c.RequestURL = append(c.RequestURL, ev.Request.URL)
				r := fetch.ContinueRequest(ev.RequestID)
				if err := r.Do(c.getExecutor(ctx)); err != nil {
					errors = append(errors, err)
				}
			}(ev)
		}
	})

	var filebyte []byte
	if err := chromedp.Run(ctx, chromedp.Tasks{
		fetch.Enable(),
		chromedp.Navigate(c.url.String()),
		chromedp.Sleep(3 * time.Second),
		chromedp.CaptureScreenshot(&filebyte),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			c.HTML, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	}); err != nil {
		return err
	}
	c.Shot = filebyte
	if len(errors) != 0 {
		return fmt.Errorf("errors: %v", errors)
	}
	return nil
}
