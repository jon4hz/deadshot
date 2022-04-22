package blockchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/logstream"
	"github.com/jon4hz/deadshot/internal/utils"
	"github.com/jon4hz/deadshot/pkg/ethutils"
	"github.com/jon4hz/deadshot/pkg/uniswap"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const significantDecimals = 6

var (
	ErrTxTimeout = errors.New("transaction timed out")
	ErrCancelNow = errors.New("canceling now")
)

// TradeDispatcher is a loop that checks if the price matches a target (considering the slippage) and executes the trade
// This function is blocking and should be run in a goroutine.
func (c *Client) TradeDispatcher(ctx context.Context, cancel context.CancelFunc, wallet *database.Wallet, trade *database.Trade, price *Price, logStream chan<- string) {
	logging.Log.WithFields(logrus.Fields{
		"token0": trade.GetToken0().GetContract(),
		"token1": trade.GetToken1().GetContract(),
	}).Info("starting TradeDispatcher")
	logStream <- logstream.Format("starting trade dispatcher", logstream.INFO)

	setNextBuyPrice(price, trade)
	setNextSellPrice(price, trade)

	price.SetHeartbeat(true)
	defer func() {
		price.Stop()
		logging.Log.WithFields(logrus.Fields{
			"token0": trade.GetToken0().GetContract(),
			"token1": trade.GetToken1().GetContract(),
		}).Info("stopping TradeDispatcher")
		logStream <- logstream.Format("stopping trade dispatcher and price feed", logstream.INFO)

		err := database.SaveTrade(trade)
		if err != nil {
			logging.Log.WithField("err", err).Error("failed to store trade in database")
		}
	}()

	var (
		gotInitPrice bool
		initPrice    *big.Int
		initTrade    *uniswap.Trade
	)

	for {
		initPrice, initTrade, gotInitPrice = getInitPrice(price, trade.GetToken0().GetDecimals())
		if gotInitPrice {
			break
		}
		logging.Log.Info("waiting for initial price")
		<-price.Heartbeat
	}

	logging.Log.WithFields(logrus.Fields{
		"initPrice": initPrice,
	}).Info("got init price")
	logStream <- logstream.Format(fmt.Sprintf("got initial price: %s", initTrade.ExecutionPrice.Invert().ToSignificant(significantDecimals)), logstream.INFO)
	trade.SetInitPrice(initPrice.String())

	var earlySellErrMsg sync.Once
	err := c.dispatchTrade(cancel, wallet, trade, price, logStream, &earlySellErrMsg)
	if err != nil {
		go func() {
			cancel()
		}()
	}
	for {
		select {
		case <-price.Heartbeat:
			err := c.dispatchTrade(cancel, wallet, trade, price, logStream, &earlySellErrMsg)
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func setNextBuyPrice(p *Price, t *database.Trade) {
	if nextB := t.GetNextBuyTarget(); nextB != nil {
		p.SetBuyAmount(nextB.GetActualAmount())
	} else {
		p.SetBuyAmount(nil)
	}
}

func setNextSellPrice(p *Price, t *database.Trade) {
	if nextS := t.GetNextSellTarget(); nextS != nil {
		p.SetSellAmount(nextS.GetActualAmount())
	} else {
		p.SetSellAmount(nil)
	}
}

func getInitPrice(price *Price, decimals uint8) (*big.Int, *uniswap.Trade, bool) {
	t := price.GetBuyTrade()
	if t == nil {
		return nil, nil, false
	}
	initPrice, gotInitPrice := getCurrenctBuyPrice(t, decimals)
	return initPrice, t, gotInitPrice
}

func getCurrenctBuyPrice(t *uniswap.Trade, decimals uint8) (*big.Int, bool) {
	ps := ethutils.ToWei(t.ExecutionPrice.Invert().Decimal(), decimals)
	if ps == nil {
		return nil, false
	}
	return ps, true
}

func getCurrentSellPrice(t *uniswap.Trade, decimals uint8) (*big.Int, bool) {
	ps := ethutils.ToWei(t.ExecutionPrice.Decimal(), decimals)
	if ps == nil {
		return nil, false
	}
	return ps, true
}

func (c *Client) dispatchTrade(cancel context.CancelFunc, wallet *database.Wallet, trade *database.Trade, price *Price, logStream chan<- string, earlySellErrMsg *sync.Once) error {
	if price.GetError() != nil {
		return nil
	}
	currentBuyTrade, currentSellTrade := price.GetTrades()
	var eg errgroup.Group
	eg.Go(func() error {
		currentBuyPrice, ok := getCurrenctBuyPrice(currentBuyTrade, trade.GetToken0().GetDecimals())
		if !ok {
			logging.Log.Error(ErrParsingPrice.Error())
			logStream <- logstream.Format(ErrParsingPrice.Error(), logstream.ERR)
			return ErrParsingPrice
		}
		for _, v := range trade.GetBuyTargets() {
			// check if the price is in the range of the target
			if !v.GetHit() && !v.GetConfirmed() && v.TriggerFunc(currentBuyPrice, v.GetPrice(), v.GetStopLoss()) {
				v.SetHit(true)
				// logStream <- logstream.Format("buy target triggered", logstream.INFO)
				err := setMissingBuyTargetInfo(v, trade.GetToken0(), currentBuyTrade.Route, trade.GetNetwork().GetWETH(), trade.GetDex().GetFeeBigInt())
				if err != nil {
					logStream <- logstream.Format(fmt.Sprintf("an error unexpected occurred: %s", err), logstream.ERR)
					return err
				}
				// SWAP
				go handleSwap(cancel, c, logStream, wallet, trade, v)
			}
		}
		return nil
	})
	eg.Go(func() error {
		currentSellPrice, ok := getCurrentSellPrice(currentSellTrade, trade.GetToken0().GetDecimals())
		if !ok {
			logging.Log.Error(ErrParsingPrice.Error())
			logStream <- logstream.Format(ErrParsingPrice.Error(), logstream.ERR)
			return nil
		}
		for _, v := range trade.GetSellTargets() {
			p := v.GetPrice()
			// check if the price is in the range of the target
			if !v.GetHit() && !v.GetConfirmed() && v.TriggerFunc(currentSellPrice, p, v.GetStopLoss()) {
				// cancel if no buy targets are hit yet
				if trade.GetBuyTargetHit() == 0 {
					earlySellErrMsg.Do(func() {
						logStream <- logstream.Format("sell target triggered but no buy target is hit yet, waiting for buy...", logstream.WARN)
						logging.Log.WithFields(logrus.Fields{"cprice": currentSellPrice.String(), "tprice": p.String()}).Warn("sell target triggered but no buy target is hit yet")
					})
					return nil
				}

				v.SetHit(true)
				// logStream <- logstream.Format("sell target triggered", logstream.INFO) // commented out, as it can flood

				err := setMissingSellTargetInfo(v, trade.GetToken1(), currentSellPrice, currentSellTrade.Route, trade.GetNetwork().GetWETH(), trade.GetDex().GetFeeBigInt())
				if err != nil {
					// if the actual sell amount is unknown, which is the case if the buy transaction is not confirmed yet
					// then we skip the error, mark the target as not hit and return nil.
					// When the next dispatcher runs, the target gets evaluated again.
					if errors.As(err, &uniswap.ErrNilAmount) {
						v.SetHit(false)
						return nil
					}
					logStream <- logstream.Format(fmt.Sprintf("an error unexpected occurred: %s", err), logstream.ERR)
					return err
				}
				// SWAP
				go handleSwap(cancel, c, logStream, wallet, trade, v)
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		logging.Log.WithField("err", err).Error("error while looping through targets")
		return err
	}
	return nil
}

// set the missing informations for the buy target.
func setMissingBuyTargetInfo(target *database.Target, token *database.Token, route *uniswap.Route, weth string, dexFee *big.Int) error {
	target.SetDefaults()

	uniToken, err := token.ToUniswap(weth)
	if err != nil {
		return fmt.Errorf("could not convert token to uniswap token: %w", err)
	}
	amount, err := uniswap.NewTokenAmount(uniToken, target.GetActualAmount())
	if err != nil {
		return fmt.Errorf("could not convert token amount to uniswap token amount: %w", err)
	}
	trade, err := uniswap.NewTrade(route, amount, uniswap.ExactInput, dexFee)
	if err != nil {
		return fmt.Errorf("could not create trade: %w", err)
	}
	slippage := int64(target.GetSlippage())
	if slippage == database.MaxSlippage {
		target.SetAmountMinMax("0")
	} else {
		min, err := trade.MinimumAmountOut(uniswap.NewPercent(big.NewInt(slippage), big.NewInt(database.MaxSlippage)))
		if err != nil {
			return fmt.Errorf("could not get minimum amount out: %w", err)
		}
		target.SetAmountMinMax(min.Raw().String())
	}

	target.SetPath(route.GetAddresses())
	target.SetExecutionPrice(trade.ExecutionPrice.Invert().Decimal())
	logging.Log.WithFields(logrus.Fields{"execution price": target.GetExecutionPrice().String()}).Info("set buy infos")
	return nil
}

// set the missing informations for the sell target.
func setMissingSellTargetInfo(target *database.Target, token *database.Token, price *big.Int, route *uniswap.Route, weth string, dexFee *big.Int) error {
	target.SetDefaults()

	if target.GetPercentageAmount() == 0 {
		mul := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(target.GetActualAmountDecimals())), nil)
		target.SetActualAmount(new(big.Int).Div(new(big.Int).Mul(target.GetAmount(), mul), price))
	}

	uniToken, err := token.ToUniswap(weth)
	if err != nil {
		return fmt.Errorf("could not convert token to uniswap token: %w", err)
	}
	amount, err := uniswap.NewTokenAmount(uniToken, target.GetActualAmount())
	if err != nil {
		return fmt.Errorf("could not convert token amount to uniswap token amount: %w", err)
	}
	trade, err := uniswap.NewTrade(route, amount, uniswap.ExactInput, dexFee)
	if err != nil {
		return fmt.Errorf("could not create trade: %w", err)
	}
	slippage := int64(target.GetSlippage())
	if slippage == database.MaxSlippage {
		target.SetAmountMinMax("0")
	} else {
		min, err := trade.MinimumAmountOut(uniswap.NewPercent(big.NewInt(slippage), big.NewInt(database.MaxSlippage)))
		if err != nil {
			return fmt.Errorf("could not get minimum amount out: %w", err)
		}
		target.SetAmountMinMax(min.Raw().String())
	}
	target.SetPath(route.GetAddresses())
	target.SetExecutionPrice(trade.ExecutionPrice.Decimal())
	logging.Log.WithFields(logrus.Fields{"execution price": target.GetExecutionPrice().String()}).Info("set sell infos")
	return nil
}

func handleSwap(
	cancel context.CancelFunc,
	c *Client,
	logStream chan<- string,
	wallet *database.Wallet,
	trade *database.Trade,
	target *database.Target,
) {
	preBal, err := database.FetchBalanceByContractAndNetworkID(trade.GetToken1().GetContract(), trade.GetNetwork().GetID())
	if err != nil {
		logging.Log.WithFields(
			logrus.Fields{
				"err": err,
			},
		).Error("failed to get pre trade balance ")
		logStream <- logstream.Format("failed to get pre trade balance", logstream.ERR)
		cancel()
		return
	}
	logging.Log.Info("pre balance: ", preBal)
	logging.Log.WithFields(
		logrus.Fields{
			"price":     target.GetPrice(),
			"amount":    target.GetActualAmount(),
			"amountmin": target.GetAmountMinMax(),
			"len route": len(target.GetPath()),
			"deadline":  target.GetDeadline(),
		},
	).Info("swapping")

	var tx *types.Transaction
	err = utils.RetryLoop(3, time.Millisecond*50, func() error {
		tx, err = c.Swap(wallet, trade, target)
		return err
	})
	if err != nil {
		logStream <- logstream.Format("could not send transaction", logstream.ERR)
		logging.Log.Error(err)
		target.SetFailed()
		cancel()
		return
	}
	logStream <- logstream.Format(fmt.Sprintf("new transaction: %s", tx.Hash().Hex()), logstream.INFO)
	target.SetTxHash(tx.Hash().Hex())
	gas, success, err := c.transactionDelegator(tx)
	if err != nil {
		logStream <- logstream.Format(err.Error(), logstream.ERR)
		logging.Log.Error(err)
		cancel()
		return
	}
	if !success {
		logStream <- logstream.Format("transaction failed", logstream.ERR)
		logging.Log.Error("transaction failed")
		target.SetFailed()
		cancel()
		return
	}

	postBal0, postBal1, err := getPostBalances(c, wallet.GetWallet(), trade.GetToken0().GetContract(), trade.GetToken1().GetContract())
	if err != nil {
		logging.Log.WithFields(
			logrus.Fields{
				"err": err,
			},
		).Error("failed to get post trade balance ")
		logStream <- logstream.Format("failed to get post trade balance", logstream.ERR)
		cancel()
		return
	}

	logging.Log.Info("post balance: ", postBal1)
	target.SetConfirmed(true)

	trade.GetToken0().SetBalance(postBal0)
	err = database.UpdateBalanceByContractAndNetworkID(trade.GetToken0().GetContract(), trade.GetNetwork().GetID(), postBal0)
	if err != nil {
		logging.Log.WithField("error", err).Error("failed to store balance")
	}

	trade.GetToken1().SetBalance(postBal1)
	err = database.UpdateBalanceByContractAndNetworkID(trade.GetToken1().GetContract(), trade.GetNetwork().GetID(), postBal1)
	if err != nil {
		logging.Log.WithField("error", err).Error("failed to store balance")
	}

	// buys
	if target.GetTargetType().Is(database.DefaultTargetTypes.GetBuy()) {
		diff := postBal1.Sub(postBal1, preBal)
		if trade.GetToken1().GetNative() {
			diff = new(big.Int).Add(diff, gas)
		}
		trade.AmountInTrade().Add(trade.AmountInTrade(), diff)
		trade.TotalBought().Add(trade.TotalBought(), diff)
		logging.Log.Info("amount bought: ", diff)
		logStream <- logstream.Format(fmt.Sprintf("amount bought: %s %s", ethutils.ShowSignificant(diff, trade.GetToken1().GetDecimals(), significantDecimals), trade.GetToken1().GetSymbol()), logstream.INFO)

		trade.IncrBuyTargetHit()

		if len(trade.GetSellTargets()) == 0 && trade.GetBuyTargetHit() == len(trade.GetBuyTargets()) {
			logStream <- logstream.Format("all buy targets hit", logstream.INFO)
			logging.Log.Info("all buy targets hit")
			cancel()
			return
		}
		
		setNextBuyPrice(price, trade)

		rebalanceSellTargets(trade)
	} else { // sells
		diff := preBal.Sub(preBal, postBal1)
		if trade.GetToken1().GetNative() {
			diff = new(big.Int).Sub(diff, gas)
		}
		trade.AmountInTrade().Sub(trade.AmountInTrade(), diff)
		logging.Log.Info("amount sold: ", diff)
		logStream <- logstream.Format(fmt.Sprintf("amount sold: %s %s", ethutils.ShowSignificant(diff, trade.GetToken1().GetDecimals(), significantDecimals), trade.GetToken1().GetSymbol()), logstream.INFO)

		// stop if the stop loss is reached
		if target.GetStopLoss() {
			logStream <- logstream.Format("stop loss reached", logstream.INFO)
			logging.Log.Info("stop loss reached")
			cancel()
			return
		}

		trade.IncrSellTargetHit()
		maxHits := len(trade.GetSellTargets())
		if trade.HasStoploss() {
			maxHits = maxHits - 1
		}
		if trade.GetSellTargetHit() >= maxHits {
			logStream <- logstream.Format("all sell targets hit, stopping now...", logstream.INFO)
			logging.Log.Info("all sell targets hit, stopping now...")
			cancel()
		}
		
		setNextSellPrice(price, trade)
	}
}

const (
	txRetryInterval = time.Millisecond * 500
	maxTxRetries    = 720 // 6 min in combination with txRetryInterval
)

// transactionDelegator returnes the gas used and if the transaction was successful.
// since the transaction receipt is not available immediately, we have to wait for it.
// the retry interval is set to 0.5 seconds and the max retries to 720 (6 min).
func (c *Client) transactionDelegator(tx *types.Transaction) (*big.Int, bool, error) {
	retryTicker := time.NewTicker(txRetryInterval)
	defer retryTicker.Stop()
	stopTimer := time.NewTimer(txRetryInterval * maxTxRetries)
	var (
		receipt *types.Receipt
		err     error
	)
	for {
		select {
		case <-stopTimer.C:
			return nil, false, ErrTxTimeout
		case <-retryTicker.C:
			if receipt == nil {
				receipt, err = c.Client.TransactionReceipt(context.Background(), tx.Hash())
				if err != nil {
					if errors.Is(err, ethereum.NotFound) {
						continue
					}
					continue
				}
				if receipt.Status == 0 { // status 0 means transaction failed
					return nil, false, nil
				}
			}
			txc, _, err := c.Client.TransactionByHash(context.Background(), tx.Hash())
			if err != nil {
				continue
			}
			totalGas := new(big.Int).Mul(txc.GasPrice(), new(big.Int).SetUint64(receipt.GasUsed))
			return totalGas, true, nil
		}
	}
}

// rebalanceRelativeSellTargets rebalances the sell targets of the trade which use a relative amount.
func rebalanceSellTargets(trade *database.Trade) {
	for _, target := range trade.GetSellTargets() {
		if target.GetHit() || target.GetConfirmed() || target.GetFailed() {
			continue
		}
		if target.GetPercentageAmount() != 0 {
			target.SetActualAmount(percentageOf(trade.TotalBought(), target.GetPercentageAmount()))
		}
		if target.GetPercentagePrice() != 0 {
			averageBuyPrice := trade.AverageBuyPrice()
			logging.Log.Debug("average buy price: ", averageBuyPrice)
			var price *big.Int
			if target.GetStopLoss() {
				price = percentagePriceSub(averageBuyPrice, target.GetPercentagePrice(), trade.GetToken0().GetDecimals())
			} else {
				price = percentagePriceAdd(averageBuyPrice, target.GetPercentagePrice(), trade.GetToken0().GetDecimals())
			}
			logging.Log.Debug("new sell target price: ", price)
			target.SetPrice(price.String())
		}
	}
}

// nolint:mnd
var decimalHundred = decimal.NewFromInt(100)

func percentagePriceAdd(buyPrice decimal.Decimal, percentage float64, decimals uint8) *big.Int {
	p := decimal.NewFromFloat(percentage)
	x := buyPrice.Mul(p).Div(decimalHundred)
	return ethutils.ToWei(buyPrice.Add(x), decimals)
}

func percentagePriceSub(buyPrice decimal.Decimal, percentage float64, decimals uint8) *big.Int {
	p := decimal.NewFromFloat(percentage)
	x := buyPrice.Mul(p).Div(decimalHundred)
	return ethutils.ToWei(buyPrice.Sub(x), decimals)
}

func percentageOf(amount *big.Int, percentage float64) *big.Int {
	p := decimal.NewFromFloat(percentage)
	a := decimal.NewFromBigInt(amount, 0)
	return a.Mul(p).Div(decimalHundred).BigInt()
}

func getPostBalances(c *Client, address, t0, t1 string) (*big.Int, *big.Int, error) {
	bal0, err := c.GetBalanceOf(address, t0)
	if err != nil {
		return nil, nil, err
	}
	bal1, err := c.GetBalanceOf(address, t1)
	if err != nil {
		return nil, nil, err
	}
	return bal0, bal1, nil
}
