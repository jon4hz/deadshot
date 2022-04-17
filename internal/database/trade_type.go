package database

import (
	"errors"
	"strings"
	"sync"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoTradeTypesFound = errors.New("no trade types found")

type TradeTypes []*TradeType

var DefaultTradeTypes TradeTypes

type TradeType struct {
	gorm.Model `yaml:"-"`
	Type       string     `yaml:"type"`
	mu         sync.Mutex `yaml:"-" gorm:"-"`
}

// Get returns the trade type.
func (t *TradeType) Get() *TradeType {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t
}

// GetName returns the type name of the TradeType.
func (t *TradeType) GetType() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Type
}

func (t *TradeType) Is(other *TradeType) bool {
	return t.Type == other.Type
}

// GetMarketTrade returns the marketTrade database object.
func (t TradeTypes) GetMarket() *TradeType {
	for _, tt := range t {
		if strings.ToLower(tt.GetType()) == "market" {
			return tt
		}
	}
	return nil
}

// GetOrderTrade returns the orderTrade database object.
func (t TradeTypes) GetOrder() *TradeType {
	for _, tt := range t {
		if strings.ToLower(tt.GetType()) == "order" {
			return tt
		}
	}
	return nil
}

// GetSnipeTrade returns the snipeTrade database object.
func (t TradeTypes) GetSnipe() *TradeType {
	for _, tt := range t {
		if strings.ToLower(tt.GetType()) == "snipe" {
			return tt
		}
	}
	return nil
}

// LoadAllTradeTypes fetches all trade types from the database and sets them as a global variable..
func LoadAllTradeTypes() error {
	result := findAllTradeTypes(&DefaultTradeTypes)
	if result.Error != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": result.Error,
		}).Error("Failed to fetch all trade types")
		return result.Error
	}
	if result.RowsAffected == 0 {
		logging.Log.Error(ErrNoTradeTypesFound)
		return ErrNoTradeTypesFound
	}
	return nil
}
