package database

import (
	"errors"
	"strings"
	"sync"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoTargetTypesFound = errors.New("no target types found")

type TargetTypes []*TargetType

var DefaultTargetTypes TargetTypes

type TargetType struct {
	gorm.Model `yaml:"-"`
	Type       string     `yaml:"type"`
	mu         sync.Mutex `yaml:"-" gorm:"-"`
}

// Is helps to compare two target types. // TODO: implement that method for trade types and amount modes too.
func (t *TargetType) Is(other *TargetType) bool {
	return t.Type == other.Type
}

// Get returns the target type.
func (t *TargetType) Get() *TargetType {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t
}

// GetType returns the type of the TargetType.
func (t *TargetType) GetType() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Type
}

// GetBuy returns the buyTarget database object.
func (t TargetTypes) GetBuy() *TargetType {
	for _, tt := range t {
		if strings.ToLower(tt.GetType()) == "buy" {
			return tt
		}
	}
	return nil
}

// GetSell returns the sellTarget database object.
func (t TargetTypes) GetSell() *TargetType {
	for _, tt := range t {
		if strings.ToLower(tt.GetType()) == "sell" {
			return tt
		}
	}
	return nil
}

// LoadAllTargetTypes fetches all trade types from the database and sets them as a global variable.
func LoadAllTargetTypes() error {
	result := findAllTargetTypes(&DefaultTargetTypes)
	if result.Error != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": result.Error,
		}).Error("Failed to fetch all target types")
		return result.Error
	}
	if result.RowsAffected == 0 {
		logging.Log.Error(ErrNoTargetTypesFound)
		return ErrNoTargetTypesFound
	}
	return nil
}
