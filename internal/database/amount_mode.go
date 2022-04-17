package database

import (
	"errors"
	"strings"
	"sync"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoAmountModeFound = errors.New("no amount mode found")

type AmountModes []*AmountMode

var DefaultAmountModes AmountModes

type AmountMode struct {
	gorm.Model `yaml:"-"`
	Type       string     `yaml:"type"`
	mu         sync.Mutex `yaml:"-" gorm:"-"`
}

// Get retruns the amount mode.
func (a *AmountMode) Get() *AmountMode {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a
}

// GetName returns the name of the AmountMode.
func (a *AmountMode) GetName() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Type
}

// GetAmountIn returns the amountIn database object.
func (a AmountModes) GetAmountIn() *AmountMode {
	for _, amountMode := range a {
		if strings.ToLower(amountMode.GetName()) == "amountin" {
			return amountMode
		}
	}
	return nil
}

// GetAmountOut returns the amountOut database object.
func (a AmountModes) GetAmountOut() *AmountMode {
	for _, amountMode := range a {
		if strings.ToLower(amountMode.GetName()) == "amountout" {
			return amountMode
		}
	}
	return nil
}

// LoadAllAmountModes fetches all amount modes from the database and sets them as a global variable..
func LoadAllAmountModes() error {
	result := findAllAmountModes(&DefaultAmountModes)
	if result.Error != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": result.Error,
		}).Error("Failed to fetch all amount modes")
		return result.Error
	}
	if result.RowsAffected == 0 {
		logging.Log.Error(ErrNoAmountModeFound)
		return ErrNoAmountModeFound
	}
	return nil
}
