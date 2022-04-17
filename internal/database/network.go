package database

import (
	"errors"
	"sync"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNoNetworksFound  = errors.New("no networks found")
	ErrNetworkNotFound  = errors.New("network not found")
	ErrNoCustomEndpoint = errors.New("no custom endpoints found")
)

type Networks []*Network

type Network struct {
	gorm.Model     `yaml:"-"`
	Endpoints      Endpoints  `yaml:"endpoints" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Tokens         []*Token   `yaml:"tokens" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Dexes          []*Dex     `yaml:"dexes" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Name           string     `yaml:"name" gorm:"unique"`
	FullName       string     `yaml:"fullName"`
	Multicall      string     `yaml:"multicall"`
	NativeCurrency string     `yaml:"nativeCurrency"`
	WETH           string     `yaml:"weth"` // represents the native currency as token
	GasLimit       uint64     `yaml:"gasLimit"`
	mu             sync.Mutex `yaml:"-" gorm:"-"`
	ChainID        uint32     `yaml:"chainId"`
	IsTestnet      bool       `yaml:"isTestnet"`
	EIP1559Enabled bool       `yaml:"eip1559Enabled"`
	Predefined     bool       `yaml:"-"`
}

// Get returns the network.
func (n *Network) Get() *Network {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n
}

func (n *Network) Testnet() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.IsTestnet
}

func (n *Network) GetID() uint {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.ID
}

func (n *Network) GetName() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.Name
}

func (n *Network) GetFullName() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.FullName
}

func (n *Network) GetChainID() uint32 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.ChainID
}

func (n *Network) GetMulticall() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.Multicall
}

func (n *Network) GetNativeCurrency() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.NativeCurrency
}

func (n *Network) NewNativeCurrencyToken() *Token {
	n.mu.Lock()
	defer n.mu.Unlock()
	t := new(Token)
	for _, token := range n.Tokens {
		if token.GetContract() == n.WETH {
			*t = *token.Get()
		}
	}
	t.SetSymbol(n.NativeCurrency)
	t.SetNative(true)
	return t
}

func (n *Network) GetGasLimit() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.GasLimit
}

func (n *Network) GetWETH() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.WETH
}

func (n *Network) GetEndpoints() Endpoints {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.Endpoints
}

func (n *Network) GetTokens() []*Token {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.Tokens
}

func (n *Network) Connectors() []*Token {
	n.mu.Lock()
	defer n.mu.Unlock()
	var connectors []*Token
	for _, token := range n.Tokens {
		if token.IsConnector() {
			connectors = append(connectors, token)
		}
	}
	return connectors
}

func (n *Network) GetDexes() []*Dex {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.Dexes
}

func (networks Networks) GetNetworkByName(name string) *Network {
	for _, network := range networks {
		if network.GetName() == name {
			return network
		}
	}
	return nil
}

// IsNativeCurrency returns true if the given currency is native to the chain.
func (n *Network) IsNativeCurrency(currency string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.NativeCurrency == currency
}

// GetCustomEndpoint returns the custom endpoint for the given network.
func (n *Network) GetCustomEndpoint() (*Endpoint, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, endpoint := range n.Endpoints {
		if endpoint.GetCustom() {
			return endpoint, true
		}
	}
	return nil, false
}

// FetchAllNetworks fetches all networks from the database.
func FetchAllNetworks(inclTestnets bool) ([]*Network, error) {
	var networks []*Network
	result := findAllNetworks(&networks)
	if result.Error != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": result.Error,
		}).Error("Failed to fetch networks")
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		logging.Log.Error(ErrNoNetworksFound)
		return nil, ErrNoNetworksFound
	}
	var actualNetworks []*Network
	if !inclTestnets {
		for _, network := range networks {
			if !network.Testnet() {
				actualNetworks = append(actualNetworks, network)
			}
		}
	} else {
		actualNetworks = networks
	}
	return actualNetworks, nil
}

func (n *Network) CreateCustomEndpoint(endpointURL string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	endpoint := &Endpoint{
		URL:    endpointURL,
		Custom: true,
	}

	// check if there is already a custom endpoint for this network
	var currentEndpoint Endpoint
	result := findCustomEndpoint(&currentEndpoint, n.Name)
	if err := result.Error; err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error while finding custom endpoint")
	}
	if result.RowsAffected > 0 {
		endpoint.ID = currentEndpoint.ID
	}

	// find the id of the network by name
	var network Network
	result = findNetworkByName(&network, n.Name)
	if err := result.Error; err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error while finding network by name")
	}
	if result.RowsAffected == 0 {
		return ErrNetworkNotFound
	}
	endpoint.NetworkID = network.ID

	if err := saveEndpoint(endpoint).Error; err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error while saving custom endpoint")
		return err
	}
	n.Endpoints = append(n.Endpoints, endpoint)
	return nil
}

func (n *Network) RemoveCustomEndpoint() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	result := deleteCustomEndpoint(n.Name)
	if result.Error != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": result.Error,
		}).Error("Error removing custom endpoint")
		return result.Error
	}
	for i, e := range n.Endpoints {
		if e.GetCustom() {
			n.Endpoints = append(n.Endpoints[:i], n.Endpoints[i+1:]...)
			break
		}
	}
	return nil
}
