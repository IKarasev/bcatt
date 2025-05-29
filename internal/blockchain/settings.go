package blockchain

import (
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/IKarasev/bcatt/internal/globals"
)

const (
	GENESIS_BLOCK_PREV byte = 0b00
)

var (
	COINBASE_START_AMOUNT int    = 1000000
	REWARD_AMOUNT         int    = 5
	COINBASE_ADDR         string = "coinbase"
	MINE_DIFF             string = "20"
	MINE_BASE             string = "115792089237316195423570985008687907853269984665640564039457584007913129639936"
	NONCE_MAX             int    = 2147483647
	NODE_NUM              int    = 3
	WALLET_NUM            int    = 1
)

func InitBcSettings(cf *globals.BcattConfig) error {
	if cf != nil {
		err := SetBcSettings(cf.Blockchain)
		return err
	}
	return ReadBcSettingsEnv()
}

func ReadBcSettingsEnv() error {
	errStr := ""
	if v := os.Getenv("BCATT_COINBASE_START"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			COINBASE_START_AMOUNT = c
		} else {
			errStr += "Failed to parse BCATT_COINBASE_START env variable\n"
		}
	}

	if v := os.Getenv("BCATT_REWARD_AMOUNT"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			REWARD_AMOUNT = c
		} else {
			errStr += "Failed to parse BCATT_REWARD_AMOUNT env variable\n"
		}
	}

	if v := os.Getenv("BCATT_MINE_DIFF"); v != "" {
		if _, ok := new(big.Int).SetString(v, 10); ok {
			MINE_DIFF = v
		} else {
			errStr += "Failed to parse BCATT_MINE_DIFF env variable\n"
		}
	}

	if v := os.Getenv("BCATT_NONCE_MAX"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			NONCE_MAX = c
		} else {
			errStr += "Failed to parse NONCE_MAX env variable\n"
		}
	}

	if v := os.Getenv("BCATT_NODE_NUM"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			NODE_NUM = c
		} else {
			errStr += "Failed to parse BCATT_NODE_NUM env variable\n"
		}
	}

	if v := os.Getenv("BCATT_WALLET_NUM"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			WALLET_NUM = c
		} else {
			errStr += "Failed to parse BCATT_WALLET_NUM env variable\n"
		}
	}

	if errStr != "" {
		return fmt.Errorf("Set Blockchain settings from env: Some settings failed to load, using default values:\n%s", errStr)
	}
	return nil
}

func SetBcSettings(cf globals.ConfigBlockchain) error {
	errStr := ""
	COINBASE_START_AMOUNT = cf.CoinbaseStart
	REWARD_AMOUNT = cf.Reward
	NONCE_MAX = cf.NonceMax
	if cf.Nodes > 0 {
		NODE_NUM = cf.Nodes
	}
	if cf.Wallets > 0 {
		WALLET_NUM = cf.Wallets
	}
	if _, ok := new(big.Int).SetString(cf.MiningDiff, 10); ok {
		MINE_DIFF = cf.MiningDiff
	} else {
		errStr += "mining_diff settings failed to parse as integer"
	}
	if errStr != "" {
		return fmt.Errorf("Set Blockchain settings from config: Some settings failed to load, using default values:\n%s", errStr)
	}
	return nil
}
