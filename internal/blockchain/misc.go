package blockchain

import (
	"context"
	"math/big"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

// GetChainID returns the chain id of the given node url.
func GetChainID(url string) (*big.Int, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"URL":   url,
			"Error": err,
		}).Error("Failed to create client")
		return nil, err
	}
	defer client.Close()

	id, err := client.ChainID(context.Background())
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"URL":   url,
			"Error": err,
		}).Error("Failed to fetch chain id")
		return nil, err
	}
	return id, nil
}

// ValidateEndpointURL validates the given node url by trying to fetch the chain id.
func ValidateEndpointURL(url string, chainID uint32) bool {
	client, err := ethclient.Dial(url)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"URL":   url,
			"Error": err,
		}).Error("Failed to create client")
		return false
	}
	defer client.Close()

	id, err := client.ChainID(context.Background())
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"URL":   url,
			"Error": err,
		}).Error("Failed to fetch chain id")
		return false
	}
	if new(big.Int).SetUint64(uint64(chainID)).Cmp(id) != 0 {
		logging.Log.WithFields(logrus.Fields{
			"URL":   url,
			"Chain": id,
		}).Warn("Chain id mismatch")
		return false
	}
	return true
}
