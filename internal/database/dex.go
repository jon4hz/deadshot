package database

import (
	"math/big"
	"sync"

	"gorm.io/gorm"
)

// Dex is the database model for a decentralized exchange.
type Dex struct {
	gorm.Model `yaml:"-"`
	Name       string `yaml:"name"`
	Router     string `yaml:"router"`
	Factory    string `yaml:"factory"`
	Fee        int64  `yaml:"fee"`
	Predefined bool   `yaml:"-"`
	// Trades     []*Trade `yaml:"-"`
	NetworkID uint `yaml:"-"`

	mu sync.Mutex `yaml:"-" gorm:"-"`
}

func NewDex(name, router, factory string, fee int64, predefined bool) *Dex {
	return &Dex{
		Name:       name,
		Router:     router,
		Factory:    factory,
		Fee:        fee,
		Predefined: predefined,
	}
}

// Get returns the dex.
func (d *Dex) Get() *Dex {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d
}

// GetName returns the name of the dex.
func (d *Dex) GetName() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Name
}

// GetRouter returns the router contract of the dex.
func (d *Dex) GetRouter() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Router
}

// GetFactory returns the factory contract of the dex.
func (d *Dex) GetFactory() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Factory
}

// GetFee returns the fee of the dex.
func (d *Dex) GetFee() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Fee
}

// GetFeeBigInt returns the fee of the dex as a *big.Int.
func (d *Dex) GetFeeBigInt() *big.Int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return big.NewInt(d.Fee)
}

// GetPredefined returns whether the dex is predefined.
func (d *Dex) GetPredefined() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Predefined
}

// SetPredefined sets whether the dex is predefined.
func (d *Dex) SetPredefined(predefined bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Predefined = predefined
}
