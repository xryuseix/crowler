package chromedp

import (
	"context"
	"sync"

	"golang.org/x/exp/slices"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func (c *ChromeDP) getExecutor(_ctx context.Context) context.Context {
	ctx := chromedp.FromContext(_ctx)
	return cdp.WithExecutor(_ctx, ctx.Target)
}

func (c *ChromeDP) Listen(ctx context.Context, loaded chan bool, req *sync.Map, errors *[]error) {
	var skipList = []network.ResourceType{"Fetch", "XHR"}
	var capList = []network.ResourceType{"Document", "Stylesheet", "Image", "Media", "Font", "Script"}
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
					c.RequestURL = append(c.RequestURL, ev.Request.URL)
				}
				r := fetch.ContinueRequest(ev.RequestID)
				if err := r.Do(c.getExecutor(ctx)); err != nil {
					*errors = append(*errors, err)
				}
			}(ev)
		case *page.EventLoadEventFired:
			go func() {
				loaded <- true
			}()
		case *network.EventRequestWillBeSent:
			go func(ev *network.EventRequestWillBeSent) {
				req.Store(ev.RequestID, true)
			}(ev)
		case *network.EventResponseReceived:
			go func(ev *network.EventResponseReceived) {
				if _, ok := req.Load(ev.RequestID); ok {
					req.Delete(ev.RequestID)
				}
			}(ev)
		}
	})
}
