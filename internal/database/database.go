package database

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	gormv2logrus "github.com/thomas-tacquet/gormv2-logrus"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultFolderPermissions = 0o755

var db *gorm.DB

var (
	ConfigDBFile string
	ConfigDir    string
	firstTime    bool
)

func init() {
	if ConfigDBFile == "" {
		folder, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}
		ConfigDir = filepath.Join(folder, "deadshot")
		ensureFolderExists(ConfigDir)

		ConfigDBFile = filepath.Join(ConfigDir, "config.db")
		firstTime = !checkIfFileExists(ConfigDBFile)
	}
}

func checkIfFileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

func ensureFolderExists(folder string) {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.Mkdir(folder, defaultFolderPermissions); err != nil {
			panic(err)
		}
	}
}

//go:embed "data/data.yml"
var defaultConfig []byte

func InitDB() error {
	gormLogger := gormv2logrus.NewGormlog(gormv2logrus.WithLogrus(logging.Log))
	gormLogger.LogMode(logger.Error)

	gormConfig := &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true,
	}
	var err error
	db, err = gorm.Open(sqlite.Open(ConfigDBFile), gormConfig)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Unable to open database")
		return err
	}

	// enable foreign key constraints
	db.Exec("PRAGMA foreign_keys = ON;")

	tables := []interface{}{
		&Token{},
		&Dex{},
		&Endpoint{},
		&Network{},
		&Misc{},
		&Wallet{},
		&TradeType{},
		&RawTarget{},
		&Target{},
		&AmountMode{},
		&TargetType{},
		&Trade{},
	}

	err = db.AutoMigrate(tables...)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Unable to migrate database")
		return err
	}

	if firstTime {
		err = createInitialData()
		if err != nil {
			return err
		}
	}

	// load all trade types
	if err = LoadAllTradeTypes(); err != nil {
		return err
	}

	// load all amount modes
	if err = LoadAllAmountModes(); err != nil {
		return err
	}

	// load all the target types
	if err = LoadAllTargetTypes(); err != nil {
		return err
	}
	return nil
}

func createInitialData() error {
	// load the default config
	type defaultConfigStruct struct {
		Networks    []*Network    `yaml:"networks"`
		TradeTypes  []*TradeType  `yaml:"tradeTypes"`
		AmountModes []*AmountMode `yaml:"amountModes"`
		TargetTypes []*TargetType `yaml:"targetTypes"`
	}
	var cfg defaultConfigStruct
	err := yaml.Unmarshal(defaultConfig, &cfg)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Unable to parse default config")
		return err
	}
	// Store the default networks and all subelements like dexes or tokens
	for _, network := range cfg.Networks {
		network.Predefined = true
		for i := range network.GetTokens() {
			network.GetTokens()[i].SetPredefined(true)
		}
		for i := range network.GetDexes() {
			network.GetDexes()[i].SetPredefined(true)
		}
		if err := saveNetwork(network).Error; err != nil {
			logging.Log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Unable to save network")
			return err
		}
	}
	// Store the default trade types
	for _, tradeType := range cfg.TradeTypes {
		if err := saveTradeType(tradeType).Error; err != nil {
			logging.Log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Unable to save trade type")
			return err
		}
	}
	// Store the default amount modes
	for _, amountMode := range cfg.AmountModes {
		if err := saveAmountMode(amountMode).Error; err != nil {
			logging.Log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Unable to save amount mode")
			return err
		}
	}
	// Store the default trade types
	for _, targetType := range cfg.TargetTypes {
		if err := saveTargetType(targetType).Error; err != nil {
			logging.Log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Unable to save target type")
			return err
		}
	}
	return nil
}

// Close closes the database connection.
// Use only when shutting down the program.
func Close() error {
	d, err := db.DB()
	if err != nil {
		return err
	}
	return d.Close()
}
