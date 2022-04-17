package config

import (
	"errors"
	"os"
	"path"
	"time"

	"github.com/jon4hz/deadshot/internal/database"

	"github.com/spf13/viper"
)

const defaultConfig = "./deadshot.yml"

var userConfig = func() string {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return path.Join(userDir, "deadshot", defaultConfig)
}

type Cfg struct {
	Testnet  bool   `yaml:"testnet"`
	Debug    bool   `yaml:"debug"`
	Keystore string `yaml:"keystore"`
	Password string `yaml:"password"`
}

var cfg Cfg

func Init() {
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := readConfigs(); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
}

func GetCfg() *Cfg {
	return &cfg
}

func readConfigs() error {
	var err error
	for _, file := range []string{defaultConfig, userConfig(), viper.GetString("config")} {
		if err = readConfig(file); err == nil {
			break
		}
	}
	return err
}

func readConfig(file string) error {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

type Config struct {
	Networks database.Networks
	Misc     *database.Misc
	Wallet   *database.Wallet
	Runtime  *RuntimeConfig
}

type RuntimeConfig struct {
	LastUnlockTime            time.Time
	WindowHeight, WindowWidth int
	Unlocked                  bool
	IncludeTestnet            bool
}

// Get returns the config or panics if not loaded.
func Get(inclTestnet bool) (*Config, error) {
	return load(inclTestnet)
}

func load(inclTestnet bool) (*Config, error) {
	var err error
	var config Config
	// load all networks
	config.Networks, err = database.FetchAllNetworks(inclTestnet)
	if err != nil {
		return nil, err
	}

	// load a new runtime config
	config.Runtime = NewRuntimeConfigWithDefaults()
	config.Runtime.IncludeTestnet = inclTestnet

	config.Wallet, err = database.FetchWallet()
	if err != nil {
		return nil, err
	}

	config.Misc, err = database.FetchMisc()
	if err != nil {
		return nil, err
	}
	return &config, config.validate()
}

func NewRuntimeConfigWithDefaults() *RuntimeConfig {
	return &RuntimeConfig{}
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("config not loaded")
	}
	if c.Networks == nil {
		return errors.New("network config not loaded")
	}
	if len(c.Networks) == 0 {
		return errors.New("no networks configured")
	}
	for _, v := range c.Networks {
		if v.GetChainID() == 0 {
			return errors.New("blockchain network " + v.GetName() + " has no chain id")
		}
		if v.GetMulticall() == "" {
			return errors.New("blockchain network " + v.GetName() + " has no multicall address")
		}
		if len(v.GetEndpoints()) == 0 {
			return errors.New("no endpoint configured for network " + v.GetName())
		}
		if len(v.GetTokens()) == 0 {
			return errors.New("no optional pairs configured for network " + v.GetName())
		}
		if len(v.GetDexes()) == 0 {
			return errors.New("no dexes configured for network " + v.GetName())
		}
	}
	return nil
}

// ReloadNetworks reloads the network config from the database.
func (c *Config) ReloadNetworks() error {
	networks, err := database.FetchAllNetworks(c.Runtime.IncludeTestnet)
	if err != nil {
		return err
	}
	c.Networks = networks
	return nil
}
