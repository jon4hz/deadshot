package multicall

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/jon4hz/geth-multicall/multicall"
	"github.com/sirupsen/logrus"
)

type Client struct {
	multicall.Multicall
}

// Init initializes the mulicall client.
func Init(rpc *ethclient.Client, contract string) (*Client, error) {
	var err error
	client, err := multicall.New(rpc, multicall.WithContractAddress(contract))
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":     err,
			"multicall": contract,
		}).Error("failed to create multicall client")
		return nil, err
	}
	return &Client{
		client,
	}, nil
}

type CallOpts struct {
	Block string
}

// call executes a web3 multicall.
func (c *Client) call(calls multicall.ViewCalls, opts *CallOpts) (*multicall.Result, error) { //nolint:unparam
	var b string

	if opts == nil {
		b = "latest"
	} else if opts.Block == "" {
		b = "latest"
	} else {
		b = opts.Block
	}

	res, err := c.Call(calls, b)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error executing web3 multicall")
	}
	return res, err
}
