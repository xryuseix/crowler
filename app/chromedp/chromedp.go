package chromedp

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"
	"xryuseix/crowler/app/config"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
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

func (c *ChromeDP) GetHTMLAndSS() error {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-cache", true),
		chromedp.WindowSize(1920, 1080),
		// chromedp.Flag("headless", false),
	)
	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel2 := chromedp.NewContext(
		allocCtx,
		// chromedp.WithDebugf(log.Printf),
	)
	for _, cancel := range []context.CancelFunc{cancel1, cancel2} {
		defer cancel()
	}

	r := sync.Map{}
	loaded := make(chan bool)
	var errors []error

	c.Listen(ctx, loaded, &r, &errors)

	if err := chromedp.Run(ctx, chromedp.Tasks{
		fetch.Enable(),
		chromedp.Navigate(c.url.String()),
	}); err != nil {
		return err
	}

	<-loaded
	timeout := time.After(time.Duration(config.Configs.Timeout.Fetch) * time.Second)
	tick := time.Tick(100 * time.Millisecond)
Loop:
	for {
		select {
		case <-timeout:
			break Loop
		case <-tick:
			isEmpty := true
			r.Range(func(key, value interface{}) bool {
				isEmpty = false
				return false
			})
			if isEmpty {
				break Loop
			}
		}
	}

	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := contentSize.Width, contentSize.Height

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(int64(width), int64(height), 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).Do(ctx)

			if err != nil {
				return err
			}

			// capture screenshot without clipping
			var quality int64 = 90
			c.Shot, err = page.CaptureScreenshot().
				WithQuality(quality).
				Do(ctx)

			if err != nil {
				return err
			}
			return nil
		}),
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
	if len(errors) != 0 {
		return fmt.Errorf("errors: %v", errors)
	}
	return nil
}
