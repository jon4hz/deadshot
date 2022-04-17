package database

import (
	"crypto/ecdsa"
	"errors"
	"sync"
	"time"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	tradeWalletPrivateKey *ecdsa.PrivateKey `gorm:"-"`
	FirstUse              bool              `gorm:"-"`
	Wallet                string
	WalletIndex           uint
	Label                 string     // label is either "trade" for wallets used for trading or "unlock" for wallets used to unlock the bot
	nonce                 uint64     `gorm:"-"`
	lastNonceUpdate       time.Time  `gorm:"-"`
	mu                    sync.Mutex `gorm:"-"`
}

var ErrWalletNotFound = errors.New("wallet not found")

func (w *Wallet) GetWallet() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Wallet
}

func (w *Wallet) NewWallet(walletAddr string, walletIndex uint, label string) error {
	wallet := &Wallet{
		Wallet:      walletAddr,
		WalletIndex: walletIndex,
		Label:       label,
	}
	wallet.ID = 1

	if err := saveWallet(wallet).Error; err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error saving wallet")
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.Wallet = walletAddr
	w.WalletIndex = walletIndex
	w.Label = label

	return nil
}

func (w *Wallet) GetWalletIndex() uint {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.WalletIndex
}

func (w *Wallet) SetPrivateKey(key *ecdsa.PrivateKey) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tradeWalletPrivateKey = key
}

// GetPrivateKey returns the private key for the wallet, make sure to load it first with wallet.Load().
func (w *Wallet) GetPrivateKey() *ecdsa.PrivateKey {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.tradeWalletPrivateKey
}

// SetNonce sets the nonce for the wallet.
func (w *Wallet) SetNonce(nonce uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastNonceUpdate = time.Now()
	w.nonce = nonce
}

// GetNonce returns the nonce for the wallet.
func (w *Wallet) GetNonce() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return int64(w.nonce)
}

// IncrementNonce increments the nonce.
func (w *Wallet) IncrementNonce() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastNonceUpdate = time.Now()
	w.nonce++
}

// LastNonceUpdate returns the time of the last nonce update.
func (w *Wallet) LastNonceUpdate() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastNonceUpdate
}

func FetchWallet() (*Wallet, error) {
	var wallet Wallet
	if result := findWallet(&wallet); result.Error != nil {
		return nil, result.Error
	}
	return &wallet, nil
}
