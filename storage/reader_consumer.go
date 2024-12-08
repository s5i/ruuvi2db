package storage

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

type RunReaderConsumerOpts struct {
	ReaderAddr   string
	QueryPeriod  time.Duration
	MaxStaleness time.Duration
	MACFilter    []string
	PushPointsF  func([]*data.Point) error
}

func RunReaderConsumer(ctx context.Context, opts *RunReaderConsumerOpts) error {
	endpoint, err := url.JoinPath("http://", opts.ReaderAddr, "data.json")
	if err != nil {
		return err
	}

	filter := map[string]bool{}
	for _, mac := range opts.MACFilter {
		filter[strings.ToUpper(mac)] = true
	}

	tick := time.NewTicker(opts.QueryPeriod)
	for {
		func() {
			resp, err := http.Get(endpoint)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var src, dst []*data.Point
			d := json.NewDecoder(resp.Body)
			if err := d.Decode(&src); err != nil {
				return
			}

			for _, p := range src {
				if len(opts.MACFilter) > 0 && !filter[strings.ToUpper(p.Address)] {
					continue
				}

				p.Timestamp = p.Timestamp.Truncate(opts.QueryPeriod)
				if p.Timestamp.Add(opts.MaxStaleness).Before(time.Now()) {
					continue
				}

				dst = append(dst, p)
			}

			if err := opts.PushPointsF(dst); err != nil {
				log.Print(err)
				return
			}
		}()

		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}
