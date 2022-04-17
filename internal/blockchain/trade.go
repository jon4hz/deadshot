package blockchain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/jon4hz/deadshot/internal/blockchain/abi/erc20"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

/* const (
	maxApprovalHex = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
) */

const bestTradesResults = 2

var (
	ErrNoPathFound     = errors.New("no path found")
	ErrNoPairsFound    = errors.New("no pairs found")
	ErrNoTradeFound    = errors.New("no trade found")
	ErrNoMidPriceFound = errors.New("no mid price found")
	ErrParsingPrice    = errors.New("failed to parse price")
)

func (c *Client) GetBestTradeExactOut(token0, token1 *database.Token, amount *big.Int, dex *database.Dex, tokens []*database.Token, maxHops int, weth string) (*uniswap.Trade, error) {
	uniPairs, err := c.genUniPairs(token0, token1, dex, tokens...)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
			"token1": token1.GetContract(),
		}).Error("failed to generate uniswap pairs")
		return nil, err
	}

	uniToken0, err := token0.ToUniswap(weth)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
		}).Error("failed to convert token0 to uniswap.Token")
		return nil, err
	}
	uniToken1, err := token1.ToUniswap(weth)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token1": token1.GetContract(),
		}).Error("failed to convert token1 to uniswap.Token")
		return nil, err
	}

	token1Amount, err := uniswap.NewTokenAmount(uniToken1, amount)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token1": token1.GetContract(),
		}).Error("failed to create token1 amount")
		return nil, err
	}

	trades, err := uniswap.BestTradeExactOut(
		uniPairs, uniToken0, token1Amount,
		&uniswap.BestTradeOptions{
			MaxNumResults: bestTradesResults,
			MaxHops:       maxHops,
			DexFee:        dex.GetFeeBigInt(),
		}, nil, nil, nil,
	)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
			"token1": token1.GetContract(),
		}).Error("failed to find best trade (exact out)")
		return nil, err
	}
	if len(trades) == 0 {
		logging.Log.WithFields(logrus.Fields{
			"error":  ErrNoTradeFound,
			"token0": token0.GetContract(),
			"token1": token1.GetContract(),
		}).Warn("no path found")
		return nil, ErrNoTradeFound
	}

	return trades[0], nil
}

func (c *Client) generatePairs(factory string, tokens ...*database.Token) (map[string]*Pair, error) {
	pairTokens, err := c.GetValidPairTokens(tokens, factory)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":   err,
			"tokens":  tokens,
			"factory": factory,
		}).Error("failed to get valid pair tokens")
		return nil, err
	}
	if len(pairTokens) == 0 {
		logging.Log.WithFields(logrus.Fields{
			"error":   ErrNoPairsFound,
			"tokens":  tokens,
			"factory": factory,
		}).Warn("no pairs found")
		return nil, err
	}

	pairs, err := c.GetPairInfo(pairTokens...)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to get pair info")
		return nil, err
	}
	if len(pairs) == 0 {
		logging.Log.Error("pair length is zero")
		return nil, err
	}
	return pairs, nil
}

func (c *Client) genUniPairs(token0, token1 *database.Token, dex *database.Dex, tokens ...*database.Token) ([]*uniswap.Pair, error) {
	tokens = append(tokens, token0, token1)

	pairs, err := c.generatePairs(dex.GetFactory(), tokens...)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to generate pairs")
		return nil, err
	}
	if len(pairs) == 0 {
		logging.Log.WithFields(logrus.Fields{
			"error": ErrNoPairsFound,
		}).Warn("no pairs found")
		return nil, ErrNoPairsFound
	}

	uniPairs := make([]*uniswap.Pair, len(pairs))
	var i int
	for _, v := range pairs {
		uniPairs[i], _, _, err = v.ToUniswap()
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to convert pair to uniswap pair")
			return nil, err
		}
		i++
	}
	return uniPairs, nil
}

func (c *Client) GetBestTradeExactIn(token0, token1 *database.Token, amount *big.Int, dex *database.Dex, tokens []*database.Token, maxHops int, weth string) (*uniswap.Trade, error) {
	uniPairs, err := c.genUniPairs(token0, token1, dex, tokens...)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
			"token1": token1.GetContract(),
			"dex":    dex.GetName(),
		}).Error("failed to generate uniswap pairs")
		return nil, err
	}
	uniToken0, err := token0.ToUniswap(weth)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
		}).Error("failed to convert token0 to uniswap.Token")
		return nil, err
	}
	uniToken1, err := token1.ToUniswap(weth)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token1": token1.GetContract(),
		}).Error("failed to convert token1 to uniswap.Token")
		return nil, err
	}

	token0Amount, err := uniswap.NewTokenAmount(uniToken0, amount)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
		}).Error("failed to create token0 amount")
		return nil, err
	}

	// uniswap.NewRoute(uniPairs,)
	trades, err := uniswap.BestTradeExactIn(
		uniPairs, token0Amount, uniToken1,
		&uniswap.BestTradeOptions{
			MaxNumResults: bestTradesResults,
			MaxHops:       maxHops,
			DexFee:        dex.GetFeeBigInt(),
		}, nil, nil, nil,
	)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
			"token1": token1.GetContract(),
			"dex":    dex.GetName(),
		}).Error("failed to find best trade (exact in)")
		return nil, err
	}
	if len(trades) == 0 {
		logging.Log.Error("no trade found")
		return nil, ErrNoTradeFound
	}
	return trades[0], nil
}

func (c *Client) GetBestOrderTrades(token0, token1 *database.Token, buyAmount, sellAmount *big.Int, dex *database.Dex, tokens []*database.Token, maxHops int, weth string) (*uniswap.Trade, *uniswap.Trade, error) {
	// Get all the onchain infos
	uniPairs, err := c.genUniPairs(token0, token1, dex, tokens...)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
			"token1": token1.GetContract(),
			"dex":    dex.GetName(),
		}).Error("failed to generate uniswap pairs")
		return nil, nil, err
	}
	uniToken0, err := token0.ToUniswap(weth)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.GetContract(),
		}).Error("failed to convert token0 to uniswap.Token")
		return nil, nil, err
	}
	uniToken1, err := token1.ToUniswap(weth)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token1": token1.GetContract(),
		}).Error("failed to convert token1 to uniswap.Token")
		return nil, nil, err
	}

	var buyTrade *uniswap.Trade
	var sellTrade *uniswap.Trade
	dexFee := dex.GetFeeBigInt()
	var eg errgroup.Group
	eg.Go(func() error {
		var err error
		buyTrade, err = bestXTrade(uniToken0, uniToken1, buyAmount, uniPairs, maxHops, dexFee)
		return err
	})
	eg.Go(func() error {
		var err error
		sellTrade, err = bestXTrade(uniToken1, uniToken0, sellAmount, uniPairs, maxHops, dexFee)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	return buyTrade, sellTrade, nil
}

func bestXTrade(token0, token1 *uniswap.Token, amount *big.Int, pairs []*uniswap.Pair, maxHops int, dexFee *big.Int) (*uniswap.Trade, error) {
	token0Amount, err := uniswap.NewTokenAmount(token0, amount)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.Address(),
		}).Error("failed to create token0 amount")
		return nil, err
	}
	// uniswap.NewRoute(uniPairs,)
	trades, err := uniswap.BestTradeExactIn(
		pairs, token0Amount, token1,
		&uniswap.BestTradeOptions{
			MaxNumResults: bestTradesResults,
			MaxHops:       maxHops,
			DexFee:        dexFee,
		}, nil, nil, nil,
	)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":  err,
			"token0": token0.Address(),
			"token1": token1.Address(),
		}).Error("failed to find best trade (exact in)")
		return nil, err
	}
	if len(trades) == 0 {
		logging.Log.Debug("no trade found")
		return nil, ErrNoTradeFound
	}
	return trades[0], nil
}

// CheckListed checks if there is an available trading route for two given tokens.
// If there is no route, chain.ErrNoTradeFound is returned so you should check against that specific error when calling this functions,
// other errors can be retured too.
func (c *Client) CheckListed(token0, token1 *database.Token, dex *database.Dex, weth string, tokens []*database.Token) (*uniswap.Trade, error) {
	trade, err := c.GetBestTradeExactIn(token0, token1, ethutils.ToWei(1, token0.GetDecimals()), dex, tokens, 5, weth)
	if err != nil {
		return nil, err
	}
	return trade, nil
}

// Swap triggers a swap of a target.
func (c *Client) Swap(wallet *database.Wallet, trade *database.Trade, target *database.Target) (*types.Transaction, error) { // make target method
	var t0, t1 *database.Token
	if target.GetTargetType().GetType() == database.DefaultTargetTypes.GetBuy().GetType() {
		t0 = trade.GetToken0()
		t1 = trade.GetToken1()
	} else {
		t0 = trade.GetToken1()
		t1 = trade.GetToken0()
	}

	// create a new router instance
	router, err := c.NewRouter(trade.GetDex().GetRouter())
	if err != nil {
		return nil, err
	}

	// convert deadline unix timestamp
	deadlineUnixTimestamp := time.Now().UTC().Unix() + int64(target.GetDeadline().Seconds())

	// create signer
	auth, err := bind.NewKeyedTransactorWithChainID(wallet.GetPrivateKey(), big.NewInt(int64(trade.GetNetwork().GetChainID())))
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to create signer")
		return nil, err
	}
	auth.GasLimit = *target.GetGasLimit()
	auth.GasPrice = target.GetGasPrice()

	if time.Since(wallet.LastNonceUpdate()) > time.Millisecond*500 {
		nonce, err := c.Client.PendingNonceAt(context.Background(), common.HexToAddress(wallet.GetWallet()))
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to get nonce")
			return nil, err
		}
		wallet.SetNonce(nonce)
	} else {
		wallet.IncrementNonce()
	}

	// Token0 is the native token, no approval necessary
	if t0.GetNative() {
		// ExactIn
		if target.GetAmountMode().GetName() == database.DefaultAmountModes.GetAmountIn().GetName() {
			auth.Value = target.GetActualAmount()
			auth.Nonce = big.NewInt(wallet.GetNonce())
			// send the exact amount
			tx, err := router.SwapExactETHForTokensSupportingFeeOnTransferTokens(auth, target.GetAmountMinMax(), target.GetPath(), common.HexToAddress(wallet.GetWallet()), big.NewInt(deadlineUnixTimestamp)) // receive a minimum amount
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error":        err,
					"type":         "SwapExactETHForTokensSupportingFeeOnTransferTokens",
					"amountIn":     target.GetActualAmount(),
					"amountOutMin": target.GetAmountMinMax(),
					"tradeWallet":  wallet.GetWallet(),
					"deadline":     target.GetDeadline(),
				}).Error("failed to swap")
				return nil, err
			}

			logging.Log.WithFields(logrus.Fields{
				"tx": tx.Hash().String(),
			}).Info("sent SwapExactETHForTokensSupportingFeeOnTransferTokens transaction")

			return tx, nil

			// ExactOut
		} else if target.GetAmountMode().GetName() == database.DefaultAmountModes.GetAmountOut().GetName() {
			auth.Value = target.GetAmountMinMax() // send a maximum amount
			auth.Nonce = big.NewInt(wallet.GetNonce())
			tx, err := router.SwapETHForExactTokens(auth, target.GetActualAmount(), target.GetPath(), common.HexToAddress(wallet.GetWallet()), big.NewInt(deadlineUnixTimestamp)) // receive the exact amount
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error":       err,
					"type":        "SwapETHForExactTokens",
					"amountInMax": target.GetAmountMinMax(),
					"amountOut":   target.GetActualAmount(),
					"tradeWallet": wallet.GetWallet(),
					"deadline":    target.GetDeadline(),
				}).Error("failed to swap")
				return nil, err
			}
			logging.Log.WithFields(logrus.Fields{
				"tx": tx.Hash().String(),
			}).Info("sent SwapETHForExactTokens transaction")
			return tx, nil
		}

		// token1 is native currency, approval is necessary
	} else if t1.GetNative() {
		// ExactIn
		if target.GetAmountMode().GetName() == database.DefaultAmountModes.GetAmountIn().GetName() {
			approved, err := c.manageApproval(
				common.HexToAddress(wallet.GetWallet()),
				common.HexToAddress(trade.GetDex().GetRouter()),
				common.HexToAddress(t0.GetContract()),
				target.GetActualAmount(),
				big.NewInt(int64(trade.GetNetwork().GetChainID())),
				wallet.GetNonce(),
				wallet.GetPrivateKey(),
			)
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to manage approval")
				return nil, err
			}

			// increment the nonce if there was an approval tx
			if approved {
				wallet.IncrementNonce()
			}
			auth.Nonce = big.NewInt(wallet.GetNonce())

			tx, err := router.SwapExactTokensForETHSupportingFeeOnTransferTokens(auth, target.GetActualAmount(), target.GetAmountMinMax(), target.GetPath(), common.HexToAddress(wallet.GetWallet()), big.NewInt(deadlineUnixTimestamp)) // receive a minimum t.GetAmount
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error":        err,
					"type":         "SwapExactTokensForETHSupportingFeeOnTransferTokens",
					"amountIn":     target.GetActualAmount(),
					"amountOutMin": target.GetAmountMinMax(),
					"tradeWallet":  wallet.GetWallet(),
					"deadline":     target.GetDeadline(),
				}).Error("failed to swap")
				return nil, err
			}
			logging.Log.WithFields(logrus.Fields{
				"tx": tx.Hash().String(),
			}).Info("sent SwapExactTokensForETHSupportingFeeOnTransferTokens transaction")
			return tx, nil

			// ExactOut
		} else if target.GetAmountMode().GetName() == database.DefaultAmountModes.GetAmountOut().GetName() {
			approved, err := c.manageApproval(
				common.HexToAddress(wallet.GetWallet()),
				common.HexToAddress(trade.GetDex().GetRouter()),
				common.HexToAddress(t0.GetContract()),
				target.GetAmountMinMax(),
				big.NewInt(int64(trade.GetNetwork().GetChainID())),
				wallet.GetNonce(),
				wallet.GetPrivateKey(),
			)
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to manage approval")
				return nil, err
			}

			// increment the nonce if there was an approval tx
			if approved {
				wallet.IncrementNonce()
			}
			auth.Nonce = big.NewInt(wallet.GetNonce())

			tx, err := router.SwapTokensForExactETH(auth, target.GetActualAmount(), target.GetAmountMinMax(), target.GetPath(), common.HexToAddress(wallet.GetWallet()), big.NewInt(deadlineUnixTimestamp)) // receive the exact amount
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error":       err,
					"type":        "SwapTokensForExactETH",
					"amountInMax": target.GetAmountMinMax(),
					"amountOut":   target.GetActualAmount(),
					"tradeWallet": wallet.GetWallet(),
					"deadline":    target.GetDeadline().Milliseconds(),
				}).Error("failed to swap")
				return nil, err
			}
			logging.Log.WithFields(logrus.Fields{
				"tx": tx.Hash().String(),
			}).Info("sent SwapTokensForExactETH transaction")
			return tx, nil
		}
		// both assets are tokens
	} else {
		// ExactIn
		if target.GetAmountMode().GetName() == database.DefaultAmountModes.GetAmountIn().GetName() {
			approved, err := c.manageApproval(
				common.HexToAddress(wallet.GetWallet()),
				common.HexToAddress(trade.GetDex().GetRouter()),
				common.HexToAddress(t0.GetContract()),
				target.GetActualAmount(),
				big.NewInt(int64(trade.GetNetwork().GetChainID())),
				wallet.GetNonce(),
				wallet.GetPrivateKey(),
			)
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to manage approval")
				return nil, err
			}

			// increment the nonce if there was an approval tx
			if approved {
				wallet.IncrementNonce()
			}
			auth.Nonce = big.NewInt(wallet.GetNonce())

			tx, err := router.SwapExactTokensForTokensSupportingFeeOnTransferTokens(auth, target.GetActualAmount(), target.GetAmountMinMax(), target.GetPath(), common.HexToAddress(wallet.GetWallet()), big.NewInt(deadlineUnixTimestamp)) // receive a minimum amount
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error":        err,
					"type":         "SwapExactTokensForTokensSupportingFeeOnTransferTokens",
					"amountIn":     target.GetActualAmount(),
					"amountOutMin": target.GetAmountMinMax(),
					"tradeWallet":  wallet.GetWallet(),
					"deadline":     target.GetDeadline(),
				}).Error("failed to swap")
				return nil, err
			}
			logging.Log.WithFields(logrus.Fields{
				"tx": tx.Hash().String(),
			}).Info("sent SwapExactTokensForTokensSupportingFeeOnTransferTokens transaction")
			return tx, nil

			// ExactOut
		} else if target.GetAmountMode().GetName() == database.DefaultAmountModes.GetAmountOut().GetName() {
			approved, err := c.manageApproval(
				common.HexToAddress(wallet.GetWallet()),
				common.HexToAddress(trade.GetDex().GetRouter()),
				common.HexToAddress(t0.GetContract()),
				target.GetAmountMinMax(),
				big.NewInt(int64(trade.GetNetwork().GetChainID())),
				wallet.GetNonce(),
				wallet.GetPrivateKey(),
			)
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error": err,
				}).Error("failed to manage approval")
				return nil, err
			}

			// increment the nonce if there was an approval tx
			if approved {
				wallet.IncrementNonce()
			}
			auth.Nonce = big.NewInt(wallet.GetNonce())

			tx, err := router.SwapTokensForExactTokens(auth, target.GetActualAmount(), target.GetAmountMinMax(), target.GetPath(), common.HexToAddress(wallet.GetWallet()), big.NewInt(deadlineUnixTimestamp)) // receive the exact amount
			if err != nil {
				logging.Log.WithFields(logrus.Fields{
					"error":       err,
					"type":        "SwapTokensForExactTokens",
					"amountInMax": target.GetAmountMinMax(),
					"amountOut":   target.GetActualAmount(),
					"tradeWallet": wallet.GetWallet(),
					"deadline":    target.GetDeadline(),
				}).Error("failed to swap")
				return nil, err
			}
			logging.Log.WithFields(logrus.Fields{
				"tx": tx.Hash().String(),
			}).Info("sent SwapTokensForExactTokens transaction")
			return tx, nil
		}
	}
	return nil, nil
}

// manageApproval checks if a token is already approved and approve it if not
// if manageApproval sent an approve tx, the function returns true and the nonce must be incremented.
func (c *Client) manageApproval(owner, spender, token common.Address, amount, chainID *big.Int, nonce int64, key *ecdsa.PrivateKey) (bool, error) {
	instance, err := erc20.NewErc20(token, c.Client)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to create erc20 instance")
		return false, err
	}

	allowance, err := instance.Allowance(nil, owner, spender)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error":   err,
			"owner":   owner.String(),
			"spender": spender.String(),
			"token":   token.String(),
		}).Error("failed to get allowance")
		return false, err
	}

	if allowance.Cmp(amount) < 0 {
		auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
		auth.Nonce = big.NewInt(nonce)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err,
			}).Error("failed to create signer")
			return false, err
		}
		tx, err := instance.Approve(auth, spender, amount)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error":   err,
				"owner":   owner.String(),
				"spender": spender.String(),
				"token":   token.String(),
			}).Error("failed to approve")
			return false, err
		}
		logging.Log.WithFields(logrus.Fields{
			"tx": tx.Hash().String(),
		}).Info("sent approve transaction")

		return true, nil
	}

	return false, nil
}
