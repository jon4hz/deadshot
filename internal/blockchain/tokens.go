package blockchain

import (
	"math/big"

	"github.com/jon4hz/deadshot/internal/blockchain/abi/erc20"
	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

func (c *Client) getTokenBalanceOf(address, contract string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	ctr := common.HexToAddress(contract)

	instance, err := erc20.NewErc20(ctr, c.Client)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"address":  address,
			"contract": contract,
			"error":    err,
		}).Error("Failed to create instance of ERC20 contract")
		return nil, err
	}

	balance, err := instance.BalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"address":  address,
			"contract": contract,
			"error":    err,
		}).Error("Failed to get balance of token")
		return nil, err
	}
	return balance, nil
}
