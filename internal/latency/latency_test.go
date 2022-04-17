package latency

import (
	"testing"
)

func TestHttpLatency(t *testing.T) {
	results := make(chan Result)
	doneC := make(chan struct{})

	urls := []string{
		"https://bsc-dataseed2.ninicoin.io",
		"https://bsc-dataseed.binance.org",
		"https://bsc-dataseed3.binance.org",
	}
	go GatherLatencies(urls, results, doneC)

	for r := range results {
		t.Log(r.URL, r.Latency)
	}
}

func TestWSLatency(t *testing.T) {
	results := make(chan Result)
	doneC := make(chan struct{})
	urls := []string{
		"wss://bsc-ws-node.nariox.org:443",
	}
	go GatherLatencies(urls, results, doneC)

	for r := range results {
		t.Log(r.URL, r.Latency)
	}
}

func TestMixedLatency(t *testing.T) {
	results := make(chan Result)
	doneC := make(chan struct{})
	urls := []string{
		"https://bsc-dataseed2.ninicoin.io",
		"https://bsc-dataseed.binance.org",
		"https://bsc-dataseed3.binance.org",
		"wss://bsc-ws-node.nariox.org:443",
	}
	go GatherLatencies(urls, results, doneC)

	for r := range results {
		t.Log(r.URL, r.Latency)
	}
}
