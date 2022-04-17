package database

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const MaxSlippage = 10000

type (
	Targets []*Target
)

// Target is the database model for a target.
type Target struct {
	gorm.Model
	// The trading path of the target.
	Path []common.Address `gorm:"-"`
	// The price of the target. Convert to *big.Int, normalized with decimals.
	Price string
	// The amount of the target. Convert to *big.Int, normalized with decimals.
	// The amount is always in the base currency of the trade.
	Amount string
	// Either the minAmountOut of maxAmountIn which is required in order to complete the trade.
	// This value is only known when the trigger price has been hit and must be set then.
	AmountMinMax string

	TxHash string
	// The actual amount of the target.
	// This value is used for the swap.
	// For a sell target, this amount is only know when the trigger price has been hit.
	ActualAmount *big.Int `gorm:"-"`

	PercentageAmount float64 `gorm:"-"`
	PercentagePrice  float64 `gorm:"-"`
	TargetTypeID     uint
	AmountModeID     uint
	AmountMode       *AmountMode `gorm:"foreignkey:AmountModeID"`
	// Slippage in percent muliplied by 100 to support up to 2 decimals.
	Slippage   *float64
	TargetType *TargetType `gorm:"foreignkey:TargetTypeID"`
	TradeID    uint
	Deadline   *time.Duration `gorm:"-"` // TODO: replace with unix timestamp and store in database
	GasLimit   *uint64
	// In WEI
	GasPrice             *big.Int   `gorm:"-"`
	mu                   sync.Mutex `gorm:"-"`
	Hit                  bool
	Confirmed            bool
	Failed               bool
	AmountDecimals       uint8
	ActualAmountDecimals uint8
	PriceDecimals        uint8
	IsStopLoss           bool
	TriggerFunc          func(*big.Int, *big.Int, bool) bool `gorm:"-"`
	ExecutionPrice       decimal.Decimal                     `gorm:"-"`
}

// NewExactTarget creates a new target. The price and the amount must be an exact value and covertable to a *big.Int.
// The amount is always in the base currency of the trade.
// It is possible to set the actual amount to nil, as it will be calculated when the trigger price is hit.
func NewExactTarget(
	price string, priceDecimals uint8,
	amount string, amountDecimals uint8,
	actualAmount *big.Int, actualAmountDecimals uint8,
	amountMode *AmountMode, targetType *TargetType,
	slippage float64, gasPrice *big.Int, stoploss bool,
	tradeID uint,
	triggerFunc func(*big.Int, *big.Int, bool) bool,
) *Target {
	return &Target{
		Model:                gorm.Model{},
		Price:                price,
		PriceDecimals:        priceDecimals,
		ActualAmount:         actualAmount,
		ActualAmountDecimals: actualAmountDecimals,
		Amount:               amount,
		AmountDecimals:       amountDecimals,
		AmountModeID:         amountMode.ID,
		AmountMode:           amountMode,
		TargetTypeID:         targetType.ID,
		TargetType:           targetType,
		Slippage:             &slippage,
		GasPrice:             gasPrice,
		IsStopLoss:           stoploss,
		TradeID:              tradeID,
		TriggerFunc:          triggerFunc,
	}
}

// NewPercentagePriceTarget create a new target where the trade is triggered as soon as the price changes by a certain percentage.
func NewPercentagePriceTarget(
	percentage float64, priceDecimals uint8,
	amount string, amountDecimals uint8,
	actualAmount *big.Int, actualAmountDecimals uint8,
	amountMode *AmountMode, targetType *TargetType,
	slippage float64, gasPrice *big.Int, stoploss bool,
	tradeID uint,
	triggerFunc func(*big.Int, *big.Int, bool) bool,
) *Target {
	return &Target{
		PercentagePrice:      percentage,
		PriceDecimals:        priceDecimals,
		ActualAmount:         actualAmount,
		ActualAmountDecimals: actualAmountDecimals,
		Amount:               amount,
		AmountDecimals:       amountDecimals,
		AmountModeID:         amountMode.ID,
		AmountMode:           amountMode,
		TargetTypeID:         targetType.ID,
		TargetType:           targetType,
		Slippage:             &slippage,
		GasPrice:             gasPrice,
		IsStopLoss:           stoploss,
		TradeID:              tradeID,
		TriggerFunc:          triggerFunc,
	}
}

// NewPercentageAmountTarget create a new target where the amount is a percentage of the total amount of the trade.
func NewPercentageAmountTarget(
	price string, priceDecimals uint8,
	percentageAmount float64, amountDecimals uint8,
	actualAmountDecimals uint8,
	amountMode *AmountMode, targetType *TargetType,
	slippage float64, gasPrice *big.Int, stoploss bool,
	tradeID uint,
	triggerFunc func(*big.Int, *big.Int, bool) bool,
) *Target {
	return &Target{
		Price:                price,
		PriceDecimals:        priceDecimals,
		ActualAmountDecimals: actualAmountDecimals,
		PercentageAmount:     percentageAmount,
		AmountDecimals:       amountDecimals,
		AmountModeID:         amountMode.ID,
		AmountMode:           amountMode,
		TargetTypeID:         targetType.ID,
		TargetType:           targetType,
		Slippage:             &slippage,
		GasPrice:             gasPrice,
		IsStopLoss:           stoploss,
		TradeID:              tradeID,
		TriggerFunc:          triggerFunc,
	}
}

// NewPercentagePriceAndAmountTarget creates a new target where the trade is triggered as soon as the price changes by a certain percentage and the amount is a percentage of the total amount of the trade.
func NewPercentagePriceAndAmountTarget(
	percentagePrice float64, priceDecimals uint8,
	percentageAmount float64, amountDecimals uint8,
	actualAmountDecimals uint8,
	amountMode *AmountMode, targetType *TargetType,
	slippage float64, gasPrice *big.Int, stoploss bool,
	tradeID uint,
	triggerFunc func(*big.Int, *big.Int, bool) bool,
) *Target {
	return &Target{
		PercentagePrice:      percentagePrice,
		PriceDecimals:        priceDecimals,
		ActualAmountDecimals: actualAmountDecimals,
		PercentageAmount:     percentageAmount,
		AmountDecimals:       amountDecimals,
		AmountModeID:         amountMode.ID,
		AmountMode:           amountMode,
		TargetTypeID:         targetType.ID,
		TargetType:           targetType,
		Slippage:             &slippage,
		GasPrice:             gasPrice,
		IsStopLoss:           stoploss,
		TradeID:              tradeID,
		TriggerFunc:          triggerFunc,
	}
}

// NewTargetWithDefaults creates a new target with default values.
func NewTargetWithDefaults() *Target {
	t := new(Target)
	t.SetDefaults()
	return t
}

// MarkStopLoss marks the target as a stop loss.
func (t *Target) MarkStopLoss() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.IsStopLoss = true
}

// GetStopLoss returns whether the target is a stop loss.
func (t *Target) GetStopLoss() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.IsStopLoss
}

// GetPrice returns the price as a *big.Int.
func (t *Target) GetPrice() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	target, _ := new(big.Int).SetString(t.Price, 10)
	return target
}

// GetPriceDecimals returns the price decimals.
func (t *Target) GetPriceDecimals() uint8 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.PriceDecimals
}

// GetAmount returns the amount as a *big.Int.
func (t *Target) GetAmount() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	amount, _ := new(big.Int).SetString(t.Amount, 10)
	return amount
}

// GetAmountDecimals returns the amount decimals.
func (t *Target) GetAmountDecimals() uint8 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.AmountDecimals
}

// GetHit returns whether the target has been hit.
func (t *Target) GetHit() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Hit
}

// GetPath returns the path.
func (t *Target) GetPath() []common.Address {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Path
}

// GetAmountMode returns the amount mode.
func (t *Target) GetAmountMode() *AmountMode {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.AmountMode
}

// GetAmountMinMax returns the amount min/max.
func (t *Target) GetAmountMinMax() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, _ := new(big.Int).SetString(t.AmountMinMax, 10)
	return a
}

// GetSlippage returns the slippage.
func (t *Target) GetSlippage() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Slippage == nil {
		return 0
	}
	return *t.Slippage
}

// GetDeadline returns the deadline.
func (t *Target) GetDeadline() *time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Deadline
}

// GetGasLimit returns the gas limit.
func (t *Target) GetGasLimit() *uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.GasLimit
}

// GetTargetType returns the target type.
func (t *Target) GetTargetType() *TargetType {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.TargetType
}

// GetConfirmed returns whether the target transaction has been confirmed.
func (t *Target) GetConfirmed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Confirmed
}

// GetTxHash returns the transaction hash.
func (t *Target) GetTxHash() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.TxHash
}

// SetConfirmed sets whether the target transaction has been confirmed.
func (t *Target) SetConfirmed(confirmed bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Confirmed = confirmed
}

// SetTxHash sets the transaction hash.
func (t *Target) SetTxHash(txHash string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TxHash = txHash
}

// SetTarget sets the price.
func (t *Target) SetPrice(target string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Price = target
}

// SetPriceDecimals sets the target decimals.
func (t *Target) SetPriceDecimals(decimals uint8) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.PriceDecimals = decimals
}

/* // SetAmount sets the amount.
func (t *Target) SetAmount(amount string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Amount = amount
} */

// SetAmountDecimals sets the amount decimals.
/* func (t *Target) SetAmountDecimals(decimals uint8) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AmountDecimals = decimals
} */

// SetHit sets the hit.
func (t *Target) SetHit(hit bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Hit = hit
}

// SetAmountMode sets the amount mode.
func (t *Target) SetAmountMode(amountMode *AmountMode) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AmountMode = amountMode
	t.AmountModeID = amountMode.ID
}

// SetTargetType sets the target type.
func (t *Target) SetTargetType(targetType *TargetType) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TargetType = targetType
	t.TargetTypeID = targetType.ID
}

// SetPath sets the path.
func (t *Target) SetPath(path []common.Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Path = path
}

// SetAmountMinMaxx sets the amount min/max.
func (t *Target) SetAmountMinMax(amountMinMax string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AmountMinMax = amountMinMax
}

// SetSlippage sets the slippage.
func (t *Target) SetSlippage(slippage *float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Slippage = slippage
}

// SetDeadline sets the deadline.
func (t *Target) SetDeadline(deadline *time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Deadline = deadline
}

// SetGasLimit sets the gas limit.
func (t *Target) SetGasLimit(gasLimit *uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.GasLimit = gasLimit
}

var (
	defaultDeadline = 20 * time.Minute
	DefaultSlippage = float64(100) // 1% (support up to 2 decimals)
	defaultGasLimit = uint64(1000000)
)

func (t *Target) SetDefaults() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Deadline == nil {
		t.Deadline = &defaultDeadline
	}
	if t.Slippage == nil || *t.Slippage == 0 {
		t.Slippage = &DefaultSlippage
	}
	if t.GasLimit == nil {
		t.GasLimit = &defaultGasLimit
	}
	if t.AmountMode == nil {
		t.AmountMode = DefaultAmountModes.GetAmountIn()
	}
	if t.TargetType == nil {
		t.TargetType = DefaultTargetTypes.GetBuy()
	}
}

func (t *Target) InvertAmountMode() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// return the workflow without doing anything if the amountMode isn't set yet
	if t.AmountMode == nil {
		return
	}
	if t.AmountMode.GetName() == DefaultAmountModes.GetAmountIn().GetName() {
		t.AmountMode = DefaultAmountModes.GetAmountOut()
	} else {
		t.AmountMode = DefaultAmountModes.GetAmountIn()
	}
}

// SwapPossible returns whether a swap is possible.
func (t *Target) MarketSwapPossible(tradeInfo *uniswap.Trade, token *Token) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tradeInfo == nil {
		return false
	}
	if reflect.DeepEqual(*tradeInfo, uniswap.Trade{}) {
		return false
	}
	if token.GetBalance().Cmp(new(big.Int)) == 0 {
		return false
	}
	if token.GetBalance().Cmp(tradeInfo.InputAmount().Raw()) < 0 {
		return false
	}
	if tradeInfo.InputAmount().Raw().Cmp(new(big.Int)) == 0 {
		return false
	}
	if tradeInfo.OutputAmount().Raw().Cmp(new(big.Int)) == 0 {
		return false
	}
	return true
}

// SetPercentageAmount sets the percentage amount.
func (t *Target) SetPercentageAmount(percentageAmount float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.PercentageAmount = percentageAmount
}

// GetActualAmount returns the actual amount.
func (t *Target) GetActualAmount() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ActualAmount
}

// SetActualAmount sets the actual amount.
func (t *Target) SetActualAmount(actualAmount *big.Int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ActualAmount = actualAmount
}

// GetActualAmountDecimals returns the actual amount decimals.
func (t *Target) GetActualAmountDecimals() uint8 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ActualAmountDecimals
}

// SetActualAmountDecimals sets the actual amount decimals.
func (t *Target) SetActualAmountDecimals(actualAmountDecimals uint8) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ActualAmountDecimals = actualAmountDecimals
}

// SetFailed marks the target as failed.
func (t *Target) SetFailed() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Failed = true
}

// GetFailed returns the failed status.
func (t *Target) GetFailed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Failed
}

// GetPercentageAmount returns the percentage amount.
func (t *Target) GetPercentageAmount() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.PercentageAmount
}

// GetPercentagePrice returns the percentage price.
func (t *Target) GetPercentagePrice() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.PercentagePrice
}

// GetExecutionPrice returns the execution price.
func (t *Target) GetExecutionPrice() decimal.Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ExecutionPrice
}

// SetExecutionPrice sets the execution price.
func (t *Target) SetExecutionPrice(executionPrice decimal.Decimal) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ExecutionPrice = executionPrice
}

// GetGasPrice returns the gas price.
func (t *Target) GetGasPrice() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.GasPrice
}

// SetGasPrice sets the gas price.
func (t *Target) SetGasPrice(gasPrice *big.Int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.GasPrice = gasPrice
}

// ExactBuyTrigger is the function used to trigger a buy.
func ExactBuyTrigger(price, target *big.Int, stopLoss bool) bool {
	if price == nil || target == nil {
		return false
	}
	// instant buys
	if target.Cmp(new(big.Int)) == 0 {
		logging.Log.Info("triggered instant buy")
		return true
	}
	if price.Cmp(target) <= 0 {
		logging.Log.WithFields(
			logrus.Fields{
				"price":  price.String(),
				"target": target.String(),
			}).Info("triggered buy")
		return true
	}
	return false
}

// ExactSellTrigger is the function used to trigger a sell.
func ExactSellTrigger(price, target *big.Int, stopLoss bool) bool {
	if price == nil || target == nil {
		return false
	}
	// instant buys
	if target.Cmp(new(big.Int)) == 0 {
		logging.Log.Debug("triggered instant buy")
		return true
	}
	if stopLoss {
		if price.Cmp(target) <= 0 {
			logging.Log.WithFields(
				logrus.Fields{
					"price":  price.String(),
					"target": target.String(),
				}).Debug("triggered stop loss")
			return true
		}
		return false
	}
	if price.Cmp(target) >= 0 {
		logging.Log.WithFields(
			logrus.Fields{
				"price":  price.String(),
				"target": target.String(),
			}).Debug("triggered sell")
		return true
	}
	return false
}

const showSignificantDigits = 6

// ViewAmount returns the amount as a string showing significant figures or a percentage value.
func (t *Target) ViewAmount() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Amount != "" {
		a, _ := new(big.Int).SetString(t.Amount, 10)
		return ethutils.ShowSignificant(a, t.AmountDecimals, showSignificantDigits)
	}
	if t.PercentageAmount != 0 {
		return fmt.Sprintf("%0.2f%%", t.PercentageAmount)
	}
	return ""
}

// ViewPrice returns the price as a string showing significant figures or a percentage value.
func (t *Target) ViewPrice() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Price != "" {
		p, _ := new(big.Int).SetString(t.Price, 10)
		return ethutils.ShowSignificant(p, t.PriceDecimals, showSignificantDigits)
	}
	if t.PercentagePrice != 0 {
		return fmt.Sprintf("%0.2f%%", t.PercentagePrice)
	}
	return ""
}
