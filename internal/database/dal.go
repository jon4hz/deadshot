package database

import (
	"gorm.io/gorm"
)

func findAllNetworks(dest *[]*Network) *gorm.DB {
	return db.Preload("Tokens").Preload("Endpoints").Preload("Dexes").Find(dest)
}

func findNetworkByName(dest *Network, networkName string) *gorm.DB {
	return db.Where("name = (?)", networkName).Find(dest)
}

func saveNetwork(n *Network) *gorm.DB {
	return db.Save(n)
}

func findTermsAndConditions(dest *Misc) *gorm.DB {
	return db.Find(dest, "id = (?)", 1)
}

func saveTermsAndConditions(toc *Misc) *gorm.DB {
	return db.Save(toc)
}

func saveEndpoint(endpoint *Endpoint) *gorm.DB {
	return db.Save(endpoint)
}

func findCustomEndpoint(dest *Endpoint, networkName string) *gorm.DB {
	subQuery := db.Select("id").Where("name = (?)", networkName).Table("networks")
	return db.Model(&Endpoint{}).Where("network_id = (?) AND custom = (?)", subQuery, true).Find(dest)
}

func deleteCustomEndpoint(networkName string) *gorm.DB {
	subQuery := db.Select("id").Where("name = (?)", networkName).Table("networks")
	return db.Unscoped().Where("network_id = (?) AND custom = (?)", subQuery, true).Delete(&Endpoint{})
}

func findWallet(dest *Wallet) *gorm.DB {
	return db.Find(dest, "id = (?)", 1)
}

func saveWallet(wallet *Wallet) *gorm.DB {
	return db.Save(wallet)
}

func findAllTradeTypes(dest *TradeTypes) *gorm.DB {
	return db.Find(dest)
}

func findAllTargetTypes(dest *TargetTypes) *gorm.DB {
	return db.Find(dest)
}

func saveTradeType(tradeType *TradeType) *gorm.DB {
	return db.Save(tradeType)
}

func findAllAmountModes(dest *AmountModes) *gorm.DB {
	return db.Find(dest)
}

func saveAmountMode(amountMode *AmountMode) *gorm.DB {
	return db.Save(amountMode)
}

func saveTargetType(targetType *TargetType) *gorm.DB {
	return db.Save(targetType)
}

func saveToken(token *Token) *gorm.DB {
	return db.Save(token)
}

func saveTrade(trade *Trade) *gorm.DB {
	return db.Save(trade)
}

func updateTokenBalance(tokenID uint, balance string) *gorm.DB {
	return db.Model(&Token{}).Where("id = (?)", tokenID).Update("balance", balance)
}

func findTokenIDByContractAndNetworkID(dest *uint, contract string, tradeID uint) *gorm.DB {
	return db.Select("id").Where("contract = (?) AND network_id = (?)", contract, tradeID).Table("tokens").Find(dest)
}

func findTokenBalanceByContractAndNetworkID(dest *string, contract string, tradeID uint) *gorm.DB {
	return db.Select("balance").Where("contract = (?) AND network_id = (?)", contract, tradeID).Table("tokens").Find(dest)
}

func findTokenByContractAndNetworkID(dest *Token, contract string, tradeID uint) *gorm.DB {
	return db.Where("contract = (?) AND network_id = (?)", contract, tradeID).Table("tokens").Find(dest)
}
