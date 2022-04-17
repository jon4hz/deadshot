package ethutils

import (
	"math/big"
	"reflect"
	"regexp"
	"strings"

	"github.com/jon4hz/deadshot/pkg/uniswap/number"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// IsValidAddress validate hex address.
func IsValidAddress(iaddress any) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address.
func IsZeroAddress(iaddress any) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// ToDecimal wei to decimals.
func ToDecimal(ivalue any, decimals uint8) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

// ToWei decimals to wei.
func ToWei(iamount any, decimals uint8) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case int:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	case *big.Int:
		amount = decimal.NewFromBigInt(v, 0)
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func ShowSignificant(amount *big.Int, decimals uint8, significantDigits uint, opt ...number.Option) string {
	if amount == nil { // prevent panic
		return ""
	}
	opts := number.New(number.WithGroupSeparator('\xA0'), number.WithRoundingMode(number.RoundHalfUp))
	opts.Apply(opt...)
	d := decimal.NewFromBigInt(amount, 1).Div(decimal.NewFromBigInt(big.NewInt(10), int32(decimals)))
	if d.LessThan(decimal.New(1, 0)) {
		significantDigits += countZerosAfterDecimalPoint(d.String())
	}
	opts.Apply(number.WithRoundingPrecision(int(significantDigits)))
	if v, err := number.DecimalRound(d, opts); err == nil {
		d = v
	}
	return number.DecimalFormat(d, opts)
}

func countZerosAfterDecimalPoint(d string) uint {
	grp := strings.Split(d, ".")
	if len(grp) != 2 {
		return 0
	}
	for i, v := range grp[1] {
		if v != '0' {
			return uint(i)
		}
	}
	return 0
}
