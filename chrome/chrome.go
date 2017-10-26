package chrome

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/knq/chromedp"
	"github.com/knq/chromedp/runner"
)

type Instance struct {
	sync.Mutex
	Context context.Context
	Cancel  context.CancelFunc
	Pool    chan *chromedp.CDP
}

func (i *Instance) Screenshot(url string) ([]byte, error) {
	var buf []byte
	client := <-i.Pool
	err := client.Run(
		i.Context,
		chromedp.Tasks{
			chromedp.Navigate(url),
			// if w don't slow down here we'll take a blank screenshot
			// since when nothing is loaded everything is ready :)
			chromedp.Sleep(time.Second / 2),
			// it's real annoying that we can't just wait for
			// the document or body to be ready here neither
			// work for any combination of chromedb query type
			// things I could find.... So we resort to this
			// heavy handed waiting for everything to be ready
			// maybe? I'm not 100% sure if it selects the first
			// or waits on all. Logging seems to suggest the later
			chromedp.WaitReady("*", chromedp.ByQuery),
			chromedp.CaptureScreenshot(&buf),
		},
	)
	i.Pool <- client
	return buf, err
}

func New(processPoolSize int) (*Instance, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var rval = &Instance{
		Context: ctx,
		Cancel:  cancel,
		Pool:    make(chan *chromedp.CDP, processPoolSize),
	}
	for i := 0; i < processPoolSize; i++ {
		log.Printf("Launching Client #%d", i)
		c, err := chromedp.New(ctx, chromedp.WithRunnerOptions(
			runner.Headless("/usr/bin/google-chrome", 9222+i),
			runner.Flag("headless", true),
			runner.Flag("disable-gpu", true),
			runner.Flag("no-first-run", true),
			runner.Flag("no-default-browser-check", true),
			runner.Flag("window-size", "1280,960"),
			runner.Flag("hide-scrollbars", true),
		)) //, chromedp.WithLog(log.Printf))
		if err != nil {
			log.Printf("Error Launching Client #%d", i)
			cancel()
			return nil, err
		}
		// Anecdotally speaking this priming business seems to help
		// avoid blank screens on the very first attempted screen
		// grab ¯\_(ツ)_/¯
		log.Printf("Priming Client #%d", i)
		err = c.Run(
			ctx,
			chromedp.Tasks{
				chromedp.Navigate("http://google.com/"),
				chromedp.Sleep(time.Second / 2),
				chromedp.WaitReady("*", chromedp.ByQuery),
			},
		)
		if err != nil {
			log.Printf("Error Priming Client #%d", i)
			cancel()
			return nil, err
		}
		rval.Pool <- c
	}
	return rval, nil
}
