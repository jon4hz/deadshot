package wallet

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39"
)

var (
	ErrSecretNotFound   = errors.New("secret not found")
	ErrEmptyTradeWallet = errors.New("trade wallet is empty")
)

func extractPrivateKey(cfg *database.Wallet, secret string) (*ecdsa.PrivateKey, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	var privateKey *ecdsa.PrivateKey

	privateKey, err := crypto.HexToECDSA(secret)
	if err == nil {
		return privateKey, nil
	}

	if bip39.IsMnemonicValid(secret) {
		if cfg.GetWallet() == "" {
			return nil, ErrEmptyTradeWallet
		}

		wallet, err := hdwallet.NewFromMnemonic(secret) //nolint:govet
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("error creating hdwallet from mnemonic")
			return nil, err
		}
		p := fmt.Sprintf("m/44'/60'/0'/0/%d", cfg.GetWalletIndex())
		path, err := hdwallet.ParseDerivationPath(p)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  p,
			}).Error("error parsing derivation path")
			return nil, err
		}
		account, err := wallet.Derive(path, false)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  p,
			}).Error("error deriving account")
			return nil, err
		}
		secret, err = wallet.PrivateKeyHex(account)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("error getting private key from account")
			return nil, err
		}
	}

	privateKey, err = crypto.HexToECDSA(secret)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error converting hex to ecdsa")
		return nil, err
	}

	return privateKey, nil
}

func getAddressFromKey(key *ecdsa.PrivateKey) (string, error) {
	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logging.Log.Error(errors.New("public key is not an ecdsa.PublicKey"))
		return "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address, nil
}

// GetAddrByIndex returns the address from the mnemonic phrase based on the index.
func GetAddrByIndex(mnemonic string, index uint) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error creating wallet from mnemonic")
		return "", err
	}
	p := fmt.Sprintf("m/44'/60'/0'/0/%d", index)
	path, err := hdwallet.ParseDerivationPath(p)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  p,
		}).Error("error parsing derivation path")
		return "", err
	}
	account, err := wallet.Derive(path, false)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error deriving account")
		return "", err
	}

	return account.Address.Hex(), nil
}

// GenerateNewMnemonic generates a new mnemonic phrase.
func GenerateNewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128) // 12 words
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error generating entropy")
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error generating mnemonic")
		return "", err
	}
	return mnemonic, nil
}
