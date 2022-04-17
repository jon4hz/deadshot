package database

import (
	"math/big"
	"strconv"
	"sync"

	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/pkg/ethutils"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RawTarget is a struct that contains the raw values of the price and amount of a target
// Currently the price and amount bust be a decimal string.
type RawTarget struct {
	gorm.Model
	Price       string
	Amount      string
	Slippage    float64
	GasPrice    *big.Int `gorm:"-"`
	ExactPrice  bool
	ExactAmount bool
	Stoploss    bool
	Skip        bool // Skip is true, if the raw target has been converted to a database.Target already by calling Target()
	TradeID     uint
	mu          sync.Mutex `gorm:"-"`
}

// NewRawTarget creates a new raw target.
func NewRawTarget(price, amount string, exactPrice, exactAmount bool, slippage float64, gasPrice *big.Int, stoploss bool) *RawTarget {
	return &RawTarget{
		Price:       price,
		Amount:      amount,
		ExactPrice:  exactPrice,
		ExactAmount: exactAmount,
		Slippage:    slippage,
		GasPrice:    gasPrice,
		Stoploss:    stoploss,
	}
}

// Target turns a raw target into a target.
func (t *RawTarget) Target(baseToken *Token, actualToken *Token, amountMode *AmountMode, targetType *TargetType, tradeID uint) *Target {
	t.mu.Lock()
	defer t.mu.Unlock()

	logging.Log.WithFields(
		logrus.Fields{
			"exactPrice":  t.ExactPrice,
			"price":       t.Price,
			"exactAmount": t.ExactAmount,
			"amount":      t.Amount,
		}).Debug("RawTarget.Target")

	t.Skip = true

	switch targetType.GetType() {
	case DefaultTargetTypes.GetBuy().GetType():
		if t.ExactPrice && t.ExactAmount {
			price := ethutils.ToWei(t.Price, baseToken.GetDecimals())
			amount := ethutils.ToWei(t.Amount, baseToken.GetDecimals())
			return NewExactTarget(price.String(), baseToken.GetDecimals(), amount.String(), baseToken.GetDecimals(), amount, actualToken.GetDecimals(), amountMode, targetType, t.Slippage, t.GasPrice, t.Stoploss, tradeID, ExactBuyTrigger)
		} else if !t.ExactAmount {
			amountP, _ := decimal.NewFromString(t.Amount)
			totalToken := decimal.NewFromBigInt(baseToken.GetBalance(), 0)
			amountD := totalToken.Mul(amountP).Div(decimal.NewFromInt(100))
			amount := amountD.BigInt()
			return NewExactTarget(t.Price, baseToken.GetDecimals(), amount.String(), baseToken.GetDecimals(), amount, actualToken.GetDecimals(), amountMode, targetType, t.Slippage, t.GasPrice, t.Stoploss, tradeID, ExactBuyTrigger)
		}

	case DefaultTargetTypes.GetSell().GetType():
		if t.ExactPrice && t.ExactAmount {
			price := ethutils.ToWei(t.Price, baseToken.GetDecimals())
			amount := ethutils.ToWei(t.Amount, baseToken.GetDecimals())
			return NewExactTarget(price.String(), baseToken.GetDecimals(), amount.String(), baseToken.GetDecimals(), nil, actualToken.GetDecimals(), amountMode, targetType, t.Slippage, t.GasPrice, t.Stoploss, tradeID, ExactSellTrigger)
		} else if t.ExactPrice && !t.ExactAmount {
			price := ethutils.ToWei(t.Price, baseToken.GetDecimals())
			amount, _ := strconv.ParseFloat(t.Amount, 64)
			return NewPercentageAmountTarget(price.String(), baseToken.GetDecimals(), amount, baseToken.GetDecimals(), actualToken.GetDecimals(), amountMode, targetType, t.Slippage, t.GasPrice, t.Stoploss, tradeID, ExactSellTrigger)
		} else if !t.ExactPrice && t.ExactAmount {
			price, _ := strconv.ParseFloat(t.Price, 64)
			amount := ethutils.ToWei(t.Amount, baseToken.GetDecimals())
			return NewPercentagePriceTarget(price, baseToken.GetDecimals(), amount.String(), baseToken.GetDecimals(), nil, actualToken.GetDecimals(), amountMode, targetType, t.Slippage, t.GasPrice, t.Stoploss, tradeID, ExactSellTrigger)
		} else if !t.ExactPrice && !t.ExactAmount {
			price, _ := strconv.ParseFloat(t.Price, 64)
			amount, _ := strconv.ParseFloat(t.Amount, 64)
			return NewPercentagePriceAndAmountTarget(price, baseToken.GetDecimals(), amount, baseToken.GetDecimals(), actualToken.GetDecimals(), amountMode, targetType, t.Slippage, t.GasPrice, t.Stoploss, tradeID, ExactSellTrigger)
		}
	}
	return nil
}
