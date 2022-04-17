package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPriceFeedInterval(t *testing.T) {
	interval := GetPriceFeedInterval("https://rpc-mainnet.maticvigil.com")
	assert.Equal(t, interval, time.Duration(170940170))
}
