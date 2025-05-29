package globals

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	ConfigWeb struct {
		Address       string `yaml:"address"`
		Port          string `yaml:"port"`
		RssUpdateTime int    `yaml:"rss_update_time"`
	}

	ConfigEmulator struct {
		OpPause   int       `yaml:"op_pause"`
		WithLog   bool      `yaml:"with_log"`
		StartUtxo StartUtxo `yaml:"start_utxo"`
	}

	ConfigBlockchain struct {
		Nodes         int                 `yaml:"nodes"`
		Wallets       int                 `yaml:"wallets"`
		CoinbaseStart int                 `yaml:"coinbase_start"`
		Reward        int                 `yaml:"reward"`
		MiningDiff    string              `yaml:"mining_diff"`
		NonceMax      int                 `yaml:"nonce_max"`
		NetMap        map[string][]string `yaml:"netmap"`
	}

	StartUtxo struct {
		Active  bool     `yaml:"active"`
		Wallets []string `yaml:"wallets"`
		All     bool     `yaml:"all"`
		Nmin    int      `yaml:"nmin"`
		Nmax    int      `yaml:"nmax"`
		Vmin    int      `yaml:"vmin"`
		Vmax    int      `yaml:"vmax"`
	}

	BcattConfig struct {
		Web        ConfigWeb        `yaml:"web"`
		Emulator   ConfigEmulator   `yaml:"emulator"`
		Blockchain ConfigBlockchain `yaml:"blockchain"`
	}
)

func ReadConfig(filePath string) (*BcattConfig, error) {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ReadConfig: Failed to read file: %s", filePath)
	}
	config := &BcattConfig{}
	err = yaml.Unmarshal(f, config)
	if err != nil {
		return nil, fmt.Errorf("ReadConfig: Failed to parse config file: %s", filePath)
	}
	return config, nil
}
