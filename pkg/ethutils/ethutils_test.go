package ethutils

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowSignificant(t *testing.T) {
	amount, _ := new(big.Int).SetString("1234567890123456789", 10)
	decimals := uint8(18)
	assert.Equal(t, "1.2345678901", ShowSignificant(amount, decimals, 10))
}
