package latency

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/jon4hz/go-httpstat"
	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
)

type Result struct {
	URL     string
	Latency int
}

// GatherLatencies gets the latency from the urls and returns the result in a channel.
func GatherLatencies(urls []string, results chan<- Result, doneC <-chan struct{}) {
	resultC := make(chan Result)
	go getLatencies(urls, resultC)
	for {
		select {
		case r, ok := <-resultC:
			if !ok {
				close(results)
				return
			}
			results <- r
		case <-doneC:
			return
		}
	}
}

func getLatencies(urls []string, resultC chan<- Result) {
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
		resultC <- Result{url, int(latency / time.Millisecond)}
	}
	close(resultC)
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
	ctx := httpstat.WithHTTPStat(context.Background(), &result)
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
