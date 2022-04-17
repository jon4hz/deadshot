package cmd

import (
	"errors"
	"io"
	"os"

	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/wallet"

	"github.com/99designs/keyring"
	"github.com/spf13/cobra"
)

var resetFlags struct {
	secret bool
	cfg    bool
	log    bool
	all    bool
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the config and secret",
	Long:  `Reset the config, remove the private key or mnemonic phrase from the secrets database and remove old log files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !resetFlags.secret && !resetFlags.cfg && !resetFlags.log && !resetFlags.all {
			return errors.New("please set at least one flag")
		}
		return reset(resetFlags.secret, resetFlags.cfg, resetFlags.log, resetFlags.all)
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&resetFlags.secret, "secret", "s", false, "Remove the secret from the secrets database")
	resetCmd.Flags().BoolVarP(&resetFlags.cfg, "config", "c", false, "Reset the config")
	resetCmd.Flags().BoolVarP(&resetFlags.log, "log", "l", false, "Reset the log")
	resetCmd.Flags().BoolVarP(&resetFlags.all, "all", "a", false, "Remove the secret and reset the config")
}

func reset(secret, cfg, log, all bool) error {
	// disable logging
	logging.Log.Out = io.Discard
	if all {
		secret = true
		cfg = true
		log = true
	}
	if secret {
		if err := wallet.InitKeystore("", ""); err != nil {
			return err
		}
		if err := wallet.RemoveSecret(); err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				return nil
			}
			return err
		}
		if err := os.RemoveAll(wallet.KeystoreFolder); err != nil {
			return err
		}
	}
	if cfg {
		if err := os.Remove(database.ConfigDBFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
	}
	if log {
		err := logging.RemoveLogs()
		if err != nil {
			return err
		}
	}
	return nil
}
