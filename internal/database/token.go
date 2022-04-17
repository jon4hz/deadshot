package database

import (
	"errors"
	"math/big"
	"sync"

	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Token represents a token with some basic informations
// must be identical to the struct internal/blockchain/multicall/token.
type Token struct {
	gorm.Model `yaml:"-"`
	Contract   string     `yaml:"contract"`
	Symbol     string     `yaml:"symbol"`
	Balance    string     `yaml:"-"` // converted to big.Int
	NetworkID  uint       `yaml:"-"`
	mu         sync.Mutex `yaml:"-" gorm:"-"`
	Decimals   uint8      `yaml:"decimals"`
	Connector  bool       `yaml:"connector"`
	Predefined bool       `yaml:"-"`

	Native bool `yaml:"native"`
}

// NewToken creates a new token.
// TODO: add field "connector".
func NewToken(address string, symbol string, decimals uint8, native bool, balance *big.Int) *Token {
	token := &Token{
		Contract: address,
		Symbol:   symbol,
		Decimals: decimals,
		Native:   native,
	}
	if balance != nil {
		token.Balance = balance.String()
	}

	return token
}

// Get returns the token.
func (t *Token) Get() *Token {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t
}

// GetContract returns the contract address.
func (t *Token) GetContract() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Contract
}

// GetSymbol returns the symbol of the token.
func (t *Token) GetSymbol() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Symbol
}

// SetSymbol sets the symbol of the token.
func (t *Token) SetSymbol(symbol string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Symbol = symbol
}

// GetDecimals returns the decimals of the token.
func (t *Token) GetDecimals() uint8 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Decimals
}

// SetDecimals sets the decimals of the token.
func (t *Token) SetDecimals(decimals uint8) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Decimals = decimals
}

// GetNative returns whether the token is native or not.
func (t *Token) GetNative() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Native
}

// SetNative sets the native flag of the token.
func (t *Token) SetNative(native bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Native = native
}

// GetBalance returns the balance of the token.
func (t *Token) GetBalance() *big.Int {
	t.mu.Lock()
	defer t.mu.Unlock()
	b, _ := new(big.Int).SetString(t.Balance, 10)
	return b
}

// SetBalance sets the balance of the tokn.
func (t *Token) SetBalance(value *big.Int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Balance = value.String()
}

// GetPredefined returns whether the token is predefined or not.
func (t *Token) GetPredefined() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Predefined
}

// SetPredefined sets the predefined flag of the token.
func (t *Token) SetPredefined(predefined bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Predefined = predefined
}

// ToUniswap converts a token to a uniswap.Token.
func (t *Token) ToUniswap(weth string) (*uniswap.Token, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	contract := t.Contract
	if t.Native {
		contract = weth
	}
	token, err := uniswap.NewToken(common.HexToAddress(contract), "", t.Symbol, t.Decimals)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// SetNetworkID sets the network id of the token.
func (t *Token) SetNetworkID(networkID uint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.NetworkID = networkID
}

// GetBalanceDecimal returns the balance of the token as a decimal.
func (t *Token) GetBalanceDecimal(decimals uint8) decimal.Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	b, _ := new(big.Int).SetString(t.Balance, 10)
	return ethutils.ToDecimal(b, decimals)
}

// Connector returns whether the token is a connector token or not.
// Connector tokens are used in the path finding.
func (t *Token) IsConnector() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Connector
}

// UpdateBalanceByContractAndNetworkID updates the balance of the token by the contract address and network id.
// The network id is the internal database id of the network. It's not related to the chain id.
func UpdateBalanceByContractAndNetworkID(contract string, networkID uint, balance *big.Int) error {
	if balance == nil {
		return nil
	}
	var tid uint
	res := findTokenIDByContractAndNetworkID(&tid, contract, networkID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil
	}
	if err := updateTokenBalance(tid, balance.String()).Error; err != nil {
		return err
	}
	return nil
}

// FetchBalance returns the balance of the token by the contract address and network id.
// The network id is the internal database id of the network. It's not related to the chain id.
func FetchBalanceByContractAndNetworkID(contract string, networkID uint) (*big.Int, error) {
	var balance string
	res := findTokenBalanceByContractAndNetworkID(&balance, contract, networkID)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("no results")
	}
	b, _ := new(big.Int).SetString(balance, 10)
	return b, nil
}

// SaveTokenUniqueByContractAndNetworkID saves the token in the database. The contract and network id are used to
// identify the token. If a token with the same contract and network id exists, nothing happens.
func SaveTokenUniqueByContractAndNetworkID(token *Token) error {
	var t Token
	res := findTokenByContractAndNetworkID(&t, token.Contract, token.NetworkID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		if err := saveToken(token).Error; err != nil {
			return err
		}
	}
	return nil
}
