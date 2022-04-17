package blockchain

import (
	"strings"
	"time"

	"github.com/jon4hz/deadshot/internal/blockchain/abi/uniswapv2router2"
	"github.com/jon4hz/deadshot/internal/blockchain/multicall"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jon4hz/web3-go/ethrpc"
	"github.com/jon4hz/web3-go/ethrpc/provider/httprpc"
	"github.com/sirupsen/logrus"
)

const Zero = "0x0000000000000000000000000000000000000000"

type Client struct {
	Client *ethclient.Client
	rpc    *ethrpc.ETH // required for the multicall package
	multic *multicall.Client
}

// NewClient initilalizes the blockchain clients.
func NewClient(node, multicallHex string) (*Client, error) {
	if strings.HasPrefix("http", node) {
		return newRetryableHTTPClient(node, multicallHex)
	}
	return newClient(node, multicallHex)
}

func newRetryableHTTPClient(node, multicallHex string) (*Client, error) {
	var err error
	c := new(Client)

	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryWaitMin = 200 * time.Millisecond
	rc.RetryWaitMax = 1000 * time.Millisecond
	rc.RetryMax = 5
	httpc := rc.StandardClient()

	rpc, err := rpc.DialHTTPWithClient(node, httpc)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
			"url":   node,
		}).Error("failed to create rpc client")
		return nil, err
	}

	c.Client = ethclient.NewClient(rpc)

	httprpc, err := httprpc.New(node)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":    err,
			"endpoint": node,
		}).Error("failed to connect to node")
		return nil, err
	}
	httprpc.SetHTTPClient(httpc)
	c.rpc, err = ethrpc.New(httprpc)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":    err,
			"endpoint": node,
		}).Error("failed to connect to node")
		return nil, err
	}

	m, err := multicall.Init(c.rpc, multicallHex)
	if err != nil {
		return nil, err
	}
	c.multic = m

	return c, nil
}

func newClient(node, multicallHex string) (*Client, error) {
	var err error
	c := new(Client)

	c.Client, err = ethclient.Dial(node)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
			"url":   node,
		}).Error("Failed to connect to node")
		return nil, err
	}

	c.rpc, err = ethrpc.NewWithDefaults(node)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":    err,
			"endpoint": node,
		}).Error("failed to connect to node")
		return nil, err
	}

	m, err := multicall.Init(c.rpc, multicallHex)
	if err != nil {
		return nil, err
	}
	c.multic = m

	return c, nil
}

// NewRouter initializes the uniswapv2 router.
func (c *Client) NewRouter(contract string) (*uniswapv2router2.Uniswapv2router2, error) {
	addr := common.HexToAddress(contract)
	router, err := uniswapv2router2.NewUniswapv2router2(addr, c.Client)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"router": contract,
		}).Error("Failed to create router")
		return nil, err
	}
	return router, nil
}

// ToUniswap converts a pair to a uniswap.Pair.
func (p Pair) ToUniswap() (*uniswap.Pair, *uniswap.Token, *uniswap.Token, error) {
	token0, err := uniswap.NewToken(common.HexToAddress(p.token0.GetContract()), "", p.token0.GetSymbol(), p.token0.GetDecimals())
	if err != nil {
		return nil, nil, nil, err
	}
	token1, err := uniswap.NewToken(common.HexToAddress(p.token1.GetContract()), "", p.token1.GetSymbol(), p.token1.GetDecimals())
	if err != nil {
		return nil, nil, nil, err
	}
	tokenAmount0, err := uniswap.NewTokenAmount(token0, p.reserve0)
	if err != nil {
		return nil, nil, nil, err
	}
	tokenAmount1, err := uniswap.NewTokenAmount(token1, p.reserve1)
	if err != nil {
		return nil, nil, nil, err
	}
	uniswapPair, err := uniswap.NewPair(common.HexToAddress(p.address), tokenAmount0, tokenAmount1)
	if err != nil {
		return nil, nil, nil, err
	}
	return uniswapPair, token0, token1, nil
}
