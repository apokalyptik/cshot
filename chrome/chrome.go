package chrome

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/knq/chromedp"
	"github.com/knq/chromedp/runner"
)

var PageWaitArtificialDelay = 600 * time.Millisecond

type Instance struct {
	sync.Mutex
	Context context.Context
	Cancel  context.CancelFunc
	Pool    chan *chromedp.CDP
	Prime   chromedp.Tasks
}

func (i *Instance) screenshotTasks(url string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		// if w don't slow down here we'll take a blank screenshot
		// since when nothing is loaded everything is ready :)
		chromedp.Sleep(PageWaitArtificialDelay),
		// it's real annoying that we can't just wait for
		// the document or body to be ready here neither
		// work for any combination of chromedb query type
		// things I could find.... So we resort to this
		// heavy handed waiting for everything to be ready
		// maybe? I'm not 100% sure if it selects the first
		// or waits on all. Logging seems to suggest the later
		chromedp.WaitReady("*", chromedp.ByQuery),
	}
}

func (i *Instance) Screenshot(url string) ([]byte, error) {
	var buf []byte
	tasks := i.screenshotTasks(url)
	client := <-i.Pool
	start := time.Now()
	// Note about speed for screenshots here...
	//
	// If you leave chromedp.WaitReady("*", chromedp.ByQuery) in the tasks
	// then you'll spend all your time waiting on that task to complete and
	// the screenshot will be very fast. If, however, you omit that WaitReady
	// task and go straight to the screenshot then you'll have a less
	// well rendered page and you'll still spend roughly the same amount of
	// time waiting only this time you will be waiting in the Capture-
	// Screenshot task. I have not yet seen a way to speed this up.  I think
	// this is just speaking to the not-production-ready-ness of chrome
	// headless mode
	err := client.Run(i.Context, tasks)
	if err != nil {
		log.Printf("client executed request, error=%s took=%s", err.Error(), time.Now().Sub(start).String())
	} else {
		log.Printf("client executed request, took=%s", time.Now().Sub(start).String())
		start = time.Now()
		if err := client.Run(i.Context, chromedp.CaptureScreenshot(&buf)); err != nil {
			log.Printf("client executed CaptureScreenshot, error=%s, took=%s", err.Error(), time.Now().Sub(start).String())
		} else {
			log.Printf("client executed CaptureScreenshot, took=%s", time.Now().Sub(start).String())
		}
	}
	go client.Shutdown(i.Context)
	return buf, err
}

func (i *Instance) Fill(processPoolSize int) {
	var n = 0
	for {
		n++
		if n >= processPoolSize*100 {
			n = 0
		}
		port := 9222 + n
		start := time.Now()
		c, err := chromedp.New(
			i.Context,
			chromedp.WithRunnerOptions(
				runner.Headless("/usr/bin/google-chrome", port),
				runner.Flag("headless", true),
				runner.Flag("disable-gpu", true),
				runner.Flag("no-first-run", true),
				runner.Flag("no-default-browser-check", true),
				runner.Flag("window-size", "1280,960"),
				runner.Flag("hide-scrollbars", true),
				runner.ExecPath("/usr/bin/google-chrome"),
			),
			// chromedp.WithLog(log.Printf), // Warning: prints base64encoded png data
		)
		if err != nil {
			log.Printf("error launching client,  error=%s, took=%s", err.Error(), time.Now().Sub(start).String())
			continue
		}
		// Anecdotally speaking this priming business seems to help
		// avoid blank screens on the very first attempted screen
		// grab ¯\_(ツ)_/¯
		if err := c.Run(i.Context, i.Prime); err != nil {
			log.Printf("error priming client, error=%s took=%s", err.Error(), time.Now().Sub(start).String())
			if c != nil {
				c.Shutdown(context.Background())
			}
			continue
		}
		log.Printf("launched client, took=%s", time.Now().Sub(start).String())

		i.Pool <- c
	}
}

func New(processPoolSize int) (*Instance, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var rval = &Instance{
		Context: ctx,
		Cancel:  cancel,
		Pool:    make(chan *chromedp.CDP, processPoolSize),
		Prime: chromedp.Tasks{
			chromedp.Navigate("data:text/html,%3Ch1%3EHello%2C%20World!%3C%2Fh1%3E"),
			chromedp.Sleep(100 * time.Millisecond),
			chromedp.WaitReady("*", chromedp.ByQuery),
		},
	}
	go rval.Fill(processPoolSize)
	return rval, nil
}
