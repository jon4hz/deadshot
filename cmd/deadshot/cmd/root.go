package cmd

import (
	"github.com/jon4hz/deadshot/internal/config"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/logging"
	log "github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/middleware/errhandler"
	"github.com/jon4hz/deadshot/internal/middleware/logger"
	"github.com/jon4hz/deadshot/internal/middleware/skip"
	"github.com/jon4hz/deadshot/internal/pipeline"
	"github.com/jon4hz/deadshot/internal/version"
	"github.com/jon4hz/deadshot/internal/wallet"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootOpts struct {
	testnet  bool
	debug    bool
	keystore string
}

var rootCmd = &cobra.Command{
	Version:          version.Version,
	TraverseChildren: true,
	Use:              "deadshot",
	Short:            "A terminal based trading bot",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if rootOpts.debug {
			logging.Log.SetLevel(logrus.DebugLevel)
			if err := startDebugAgents(); err != nil {
				return err
			}
		}
		wallet.SetKeyStoreBackend(rootOpts.keystore)
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := log.SetFile(); err != nil {
			return err
		}
		if err := database.InitDB(); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return root()
	},
}

func init() {
	cobra.OnInitialize(config.Init)

	rootCmd.Flags().BoolVar(&rootOpts.testnet, "testnet", false, "use testnet")
	rootCmd.Flags().BoolVar(&rootOpts.debug, "debug", false, "enable debug mode")
	rootCmd.Flags().StringVarP(&rootOpts.keystore, "keystore", "k", "auto", "Set the keystore. Available: auto, file")

	viper.BindPFlag("testnet", rootCmd.Flags().Lookup("testnet"))
	viper.BindPFlag("debug", rootCmd.Flags().Lookup("debug"))
	viper.BindPFlag("keystore", rootCmd.Flags().Lookup("keystore"))

	rootCmd.AddCommand(
		resetCmd,
		logCmd,
		uitestCmd,
		versionCmd,
	)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func root() error {
	c, err := config.Get(rootOpts.testnet)
	if err != nil {
		return err
	}
	ctx := context.New(c, config.GetCfg())
	for _, pipe := range pipeline.NewPipeline() {
		if err := skip.Maybe(
			pipe,
			logger.Log(
				pipe.String(),
				errhandler.Handle(pipe.Run),
			),
		)(ctx); err != nil {
			return err
		}
	}
	return nil
}
