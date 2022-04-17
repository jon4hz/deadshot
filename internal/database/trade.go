package database

import (
	"math/big"
	"sync"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Trade is the database model for a trade.
type Trade struct {
	gorm.Model
	RawBuyTargets  []*RawTarget `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RawSellTargets []*RawTarget `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	BuyTargets     Targets      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	SellTargets    Targets      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	InitPrice      string       // convert to *big.Int, normalized with decimals
	Token0         *Token       `gorm:"foreignkey:Token0ID"`
	Token0ID       uint
	Token1         *Token `gorm:"foreignkey:Token1ID"`
	Token1ID       uint
	buysHit        int        `gorm:"-"`
	sellsHit       int        `gorm:"-"`
	TradeType      *TradeType `gorm:"foreignkey:TradeTypeID"`
	TradeTypeID    uint
	Endpoint       *Endpoint `gorm:"foreignkey:EndpointID"`
	EndpointID     uint
	Network        *Network `gorm:"foreignkey:NetworkID"`
	NetworkID      uint
	Dex            *Dex `gorm:"foreignkey:DexID"`
	DexID          uint
	// the amount of tokens which have been bought and not sold yet
	amountInTrade *big.Int `gorm:"-"`
	// the amount of tokens which have been bought
	totalBought *big.Int `gorm:"-"`
	Failed      bool
	hasStoploss bool `gorm:"-"`
	// mutex
	mu sync.Mutex `gorm:"-"`
}

// NewTrade creates a new trade.
func NewTrade(token0, token1 *Token, buyTargets, sellTargets Targets, tradeType *TradeType, endpoint *Endpoint, network *Network, dex *Dex) *Trade {
	var hasSL bool
	for _, t := range sellTargets {
		if t.GetStopLoss() {
			hasSL = true
			break
		}
	}

	return &Trade{
		Token0:      token0,
		Token0ID:    token0.ID,
		Token1:      token1,
		Token1ID:    token1.ID,
		BuyTargets:  buyTargets,
		SellTargets: sellTargets,
		hasStoploss: hasSL,
		TradeType:   tradeType,
		TradeTypeID: tradeType.ID,
		Endpoint:    endpoint,
		EndpointID:  endpoint.ID,
		Network:     network,
		NetworkID:   network.ID,
		Dex:         dex,
		DexID:       dex.ID,
		Failed:      false,
	}
}

// GetDex returns the dex for the trade.
func (t *Trade) GetDex() *Dex {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Dex
}

// GetNetwork returns the network for the trade.
func (t *Trade) GetNetwork() *Network {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Network
}

// GetToken0 returns the token0 for the trade.
func (t *Trade) GetToken0() *Token {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Token0
}

// GetToken1 returns the token1 for the trade.
func (t *Trade) GetToken1() *Token {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Token1
}

// GetEndpoint returns the endpoint for the trade.
func (t *Trade) GetEndpoint() *Endpoint {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Endpoint
}

// InvertTokens inverts the tokens for the trade.
func (t *Trade) InvertTokens() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Token0, t.Token1 = t.Token1, t.Token0
}

// SetToken0 sets the token0 for the trade.
func (t *Trade) SetToken0(token *Token) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Token0 = token
}

// SetToken1 sets the token1 for the trade.
func (t *Trade) SetToken1(token *Token) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Token1 = token
}

// SetTargets sets the targets for the trade.
// Required for the market trades, otherwise not.
func (t *Trade) SetBuyTargets(targets []*Target) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.BuyTargets = targets
}

// GetBuyTargets returns the buy targets for the trade.
func (t *Trade) GetBuyTargets() []*Target {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.BuyTargets
}

// GetNextBuyTarget returns the first buy target that has not been hit.
// If all buy targets have been hit, it returns nil.
func (t *Trade) GetNextBuyTarget() *Target {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, target := range t.BuyTargets {
		if !target.Hit {
			return target
		}
	}
	return nil
}

// GetSellTargets returns the sell targets for the trade.
func (t *Trade) GetSellTargets() []*Target {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.SellTargets
}

// GetNextSellTarget returns the first sell target that has not been hit.
// If all sell targets have been hit, it returns it returns nil.
func (t *Trade) GetNextSellTarget() *Target {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, target := range t.SellTargets {
		if !target.Hit {
			return target
		}
	}
	return nil
}

// GetInitPrice returns the init price for the trade.
func (t *Trade) GetInitPrice() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	price, _ := new(big.Int).SetString(t.InitPrice, 10)
	return price
}

// SetInitPrice sets the init price for the trade.
func (t *Trade) SetInitPrice(price string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.InitPrice = price
}

// AmountInTrade returns the amount in trade for the trade.
func (t *Trade) AmountInTrade() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.amountInTrade == nil {
		t.amountInTrade = big.NewInt(0)
	}
	return t.amountInTrade
}

// TotalBought returns the total amount that has been bought.
func (t *Trade) TotalBought() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.totalBought == nil {
		t.totalBought = big.NewInt(0)
	}
	return t.totalBought
}

// AverageBuyPrice returns the average buy price for the trade.
func (t *Trade) AverageBuyPrice() decimal.Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.totalBought == nil || t.amountInTrade == nil {
		return decimal.Zero
	}
	// Get the execution price from all buy targets and return the average.
	totalExecutionPrice := decimal.Zero
	var counter int64
	for _, target := range t.BuyTargets {
		if target.Hit {
			totalExecutionPrice = totalExecutionPrice.Add(target.ExecutionPrice)
			counter++
		}
	}
	return totalExecutionPrice.Div(decimal.NewFromInt(counter))
}

// HasStoploss returns whether the trade has a stoploss target.
func (t *Trade) HasStoploss() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.hasStoploss
}

// IncrSellTargetHit increments the sell target hit counter.
func (t *Trade) IncrSellTargetHit() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sellsHit++
}

// GetSellTargetHit returns the sell target hit counter.
func (t *Trade) GetSellTargetHit() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sellsHit
}

// IncrBuyTargetHit increments the buy target hit counter.
func (t *Trade) IncrBuyTargetHit() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buysHit++
}

// GetBuyTargetHit returns the buy target hit counter.
func (t *Trade) GetBuyTargetHit() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buysHit
}

// SaveTrade saves the trade to the database.
func SaveTrade(t *Trade) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := saveTrade(t).Error; err != nil {
		return err
	}
	return nil
}
