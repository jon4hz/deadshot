package wallet

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/99designs/keyring"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39"
)

const (
	defaultServiceName     = "deadshot.wallet.eth"
	defaultKeyringName     = "deadshot"
	keystoreFileIdentifier = "file"
)

var (
	backend         string
	password        string
	requirePassword bool
	keystore        keyring.Keyring
	KeystoreFolder  string
)

var (
	ErrWrongPassword         = errors.New("aes.KeyUnwrap(): integrity check failed.") // nolint:revive
	ErrKeyringNotInitialized = errors.New("keyring not initialized")
)

func init() {
	KeystoreFolder = filepath.Join(database.ConfigDir, "keystore")
}

func SetKeyStoreBackend(store string) {
	switch store {
	case keystoreFileIdentifier:
		requirePassword = true
	}
	backend = store
}

func InitKeystore(password, ksBackend string) error {
	if ksBackend == "" {
		ksBackend = backend
	}
	var backends []keyring.BackendType
	switch ksBackend {
	case keystoreFileIdentifier:
		backends = []keyring.BackendType{keyring.FileBackend}
		requirePassword = true
	default:
		backends = keyring.AvailableBackends()
		if len(backends) == 0 {
			return errors.New("no keyring backends available")
		}
		if len(backends) > 1 {
			for i, b := range backends {
				// HACK: manually check if pass is actually available and if not, remove it from the list
				if b == keyring.PassBackend {
					if !checkPassKeystoreAvailable() {
						backends = append(backends[:i], backends[i+1:]...)
					}
				}
			}
			logging.Log.WithField("keystores", backends).Info("available keystore backends")
		}
		if len(backends) == 1 {
			if backends[0] == keyring.FileBackend {
				requirePassword = true
			}
			logging.Log.WithField("keystore", backends[0]).Warn("only one keystore is available")
		}
	}
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:        backends,
		ServiceName:            defaultKeyringName,
		KeychainSynchronizable: false,
		FileDir:                KeystoreFolder,
		FilePasswordFunc: func(s string) (string, error) {
			return password, nil
		},
	})
	if err != nil {
		return err
	}
	keystore = kr
	return nil
}

func checkPassKeystoreAvailable() bool {
	const passCmd = "pass"
	_, err := exec.LookPath(passCmd)
	return err == nil
}

func IsInitialized() bool {
	return keystore != nil
}

func RequirePassword() bool {
	return requirePassword
}

// CheckKeystore checks whether the keystore already exists and if the keystore requires a password.
func CheckKeystore() (bool, error) {
	if backend == keystoreFileIdentifier {
		return CheckFileKeystore()
	}
	return checkAutoKeystore()
}

func CheckFileKeystore() (bool, error) {
	_, err := os.Stat(path.Join(KeystoreFolder, defaultServiceName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func checkAutoKeystore() (bool, error) {
	_, err := getSecret()
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func Load(cfg *database.Wallet) error {
	secret, err := getSecret()
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return ErrSecretNotFound
		}
		return err
	}

	key, err := extractPrivateKey(cfg, secret)
	if err != nil {
		return err
	}

	cfg.SetPrivateKey(key)

	address, err := getAddressFromKey(key)
	if err != nil {
		return err
	}
	cfg.NewWallet(address, cfg.GetWalletIndex(), "trade")

	return nil
}

func getSecret() (string, error) {
	item, err := keystore.Get(defaultServiceName)
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

// Set sets the secret in the keystore.
// If the secret already exists, it will be overwritten.
// Set will also set the private key in the wallet config.
func Set(secret string, tradeWalletIndex uint, cfg *database.Wallet) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	err := setSecret(secret)
	if err != nil {
		return err
	}

	var address string
	if bip39.IsMnemonicValid(secret) {
		addr, err := GetAddrByIndex(secret, tradeWalletIndex) //nolint:govet
		if err != nil {
			return err
		}
		address = addr
	} else {
		privateKey, err := crypto.HexToECDSA(secret) //nolint:govet
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("error converting hex to ecdsa")
			return err
		}
		addr, err := getAddressFromKey(privateKey)
		if err != nil {
			return err
		}
		address = addr
	}

	err = cfg.NewWallet(address, tradeWalletIndex, "trade")
	if err != nil {
		return err
	}
	return nil
}

func setSecret(secret string) error {
	return keystore.Set(keyring.Item{
		Key:  defaultServiceName,
		Data: []byte(secret),
	})
}

// SearchMnemnoic searches the  secrets database for a mnemonic phrase.
func SearchMnemonic() (string, bool, error) {
	secret, err := getSecret()
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("error searching for secret")
		return "", false, err
	}
	if !bip39.IsMnemonicValid(secret) {
		return "", false, nil
	}
	return secret, true, nil
}

func RemoveSecret() error {
	if keystore == nil {
		return ErrKeyringNotInitialized
	}
	return keystore.Remove(defaultServiceName)
}
