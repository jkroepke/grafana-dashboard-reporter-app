package chrome

import (
	"fmt"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/mahendrapaipuri/grafana-dashboard-reporter-app/pkg/plugin/config"
	"golang.org/x/net/context"
)

/*
	This file contains chromedp package related helper functions.
	Sources:
		- https://github.com/chromedp/chromedp/issues/1044
		- https://github.com/chromedp/chromedp/issues/431#issuecomment-592950397
		- https://github.com/chromedp/chromedp/issues/87
		- https://github.com/chromedp/examples/tree/master
*/

// PDFOptions contains the templated HTML Body, Header and Footer strings
type PDFOptions struct {
	Header string
	Body   string
	Footer string

	Orientation string
}

type Instance interface {
	NewTab(logger log.Logger, conf *config.Config) *Tab
	Name() string
	Close(logger log.Logger)
}

// enableLifeCycleEvents enables the chromedp life cycle events
func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to enable page: %w", err)
		}

		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to enable lifecycle events: %w", err)
		}

		return nil
	}
}

// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//
//	init, DOMContentLoaded, firstPaint,
//	firstContentfulPaint, firstImagePaint,
//	firstMeaningfulPaintCandidate,
//	load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
//
// This is not super reliable, I've already found incidental cases where
// networkIdle was sent before load. It's probably smart to see how
// puppeteer implements this exactly.
func waitFor(eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		ch := make(chan struct{})
		cctx, cancel := context.WithCancel(ctx)
		chromedp.ListenTarget(cctx, func(ev interface{}) {
			switch e := ev.(type) {
			case *page.EventLifecycleEvent:
				if e.Name == eventName {
					cancel()
					close(ch)
				}
			}
		})
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// SetHeaders returns a task list that sets the passed headers.
func setHeaders(headers map[string]any) chromedp.Tasks {
	if headers == nil {
		return chromedp.Tasks{}
	}

	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(headers),
	}
}