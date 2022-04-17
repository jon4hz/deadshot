package endpoint

import (
	ctx "context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"

	chain "github.com/jon4hz/deadshot/internal/blockchain"

	"github.com/jon4hz/go-httpstat"
	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
)

var (
	_ modules.Piper    = &Pipe{}
	_ modules.Canceler = &Pipe{}
)

type Pipe struct {
	c      ctx.Context
	cancel ctx.CancelFunc
}

func (p Pipe) String() string { return "endpoint" }

func (p *Pipe) CancelFunc() ctx.CancelFunc {
	p.c, p.cancel = ctx.WithCancel(ctx.Background())
	return p.cancel
}

func (p *Pipe) Run(ctx *context.Context) error {
	results := make([]context.LatencyResult, 0)

	resultC := make(chan context.LatencyResult)
	doneC := make(chan struct{})

	endpoints := ctx.Network.GetEndpoints().GetUrls()
	custom, ok := ctx.Network.GetCustomEndpoint()
	if ok {
		endpoints = []string{custom.GetURL()}
	}

	go getLatencies(endpoints, resultC, doneC)

loop:
	for {
		select {
		case r, ok := <-resultC:
			if !ok {
				close(ctx.LatencyResultChan)
				break loop
			}
			results = append(results, r)
			ctx.LatencyResultChan <- r

		case <-ctx.LatencyResultDone:
			return nil

		case <-doneC:
			break loop
		}
	}

	if len(results) == 0 {
		return errors.New("could not connect to any endpoint")
	}
	l := int(1e5)
	var u string
	for _, v := range results {
		if v.Latency < l {
			l = v.Latency
			u = v.URL
		}
	}

	// set all values to the pipeline or cancel if context is not valid anymore
	select {
	case <-p.c.Done():
		return nil
	default:
	}

	ctx.Endpoint = ctx.Network.GetEndpoints().GetEndpointByURL(u)
	ctx.BestLatency = l

	client, err := chain.NewClient(ctx.Endpoint.GetURL(), ctx.Network.GetMulticall())
	if err != nil {
		return err
	}
	ctx.Client = client

	return nil
}

func getLatencies(urls []string, resultC chan<- context.LatencyResult, doneC chan<- struct{}) {
	for _, url := range urls {
		// Create a httpstat powered context
		var result *httpstat.Result
		var err error
		if strings.HasPrefix(url, "http") {
			result, err = getHTTPLatency(url)
		} else if strings.HasPrefix(url, "ws") {
			result, err = getWsLatency(url)
		}
		if err != nil {
			continue
		}
		var latency time.Duration
		latency += result.TCPConnection
		latency += result.TLSHandshake
		latency += result.Connect
		resultC <- context.LatencyResult{
			URL:     url,
			Latency: int(latency / time.Millisecond),
		}
	}
	close(doneC)
}

func getHTTPLatency(url string) (*httpstat.Result, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
			"url":   url,
		}).Error("Failed to create a new HTTP request")
		return nil, err
	}
	var result httpstat.Result
	ctx := httpstat.WithHTTPStat(req.Context(), &result)
	req = req.WithContext(ctx) // Send request by default HTTP client
	client := new(http.Client)
	defer client.CloseIdleConnections()
	res, err := client.Do(req)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
			"url":   url,
		}).Error("Failed to send a HTTP request")
		return nil, err
	}
	result.End(time.Now())
	defer res.Body.Close()
	return &result, nil
}

func getWsLatency(url string) (*httpstat.Result, error) {
	var result httpstat.Result
	ctx := httpstat.WithHTTPStat(ctx.Background(), &result)
	_, _, err := websocket.Dial(ctx, url, nil) //nolint:bodyclose
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
			"url":   url,
		}).Error("Failed to dial a websocket")
		return nil, err
	}
	result.End(time.Now())
	return &result, nil
}
