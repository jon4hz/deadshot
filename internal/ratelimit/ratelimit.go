package ratelimit

import "time"

// common rate limits for the rpc endpoints per millisecond.
var rateLimits = map[string]float64{
	"https://rpc-mainnet.maticvigil.com": 0.0117, // 700 per minute to be precise
}

func GetPriceFeedInterval(url string) time.Duration {
	limit, ok := rateLimits[url]
	if !ok {
		return 0
	}
	return time.Duration(float64(1)/limit*1000000) * 2 // multiply by 2 because the price feed requires two requests per interval
}
