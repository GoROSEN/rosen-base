package config

type BlockChainTokenConfig struct {
	Name            string `yaml:"name"`
	AccountType     string `yaml:"accountType"`
	ContractAddress string `yaml:"contractAddress"`
	AutoCreate      bool   `yaml:"autoCreate"`
	Decimals        uint64 `yaml:"decimals"`
}

type BlockchainConfig struct {
	Name              string                   `yaml:"name"`
	Endpoint          string                   `yaml:"endpoint"`
	WsEndpoint        string                   `yaml:"ws_endpoint"`
	Funder            string                   `yaml:"funder"`
	UnlockPhrase      string                   `yaml:"unlockPhrase"`
	RateLimit         int                      `yaml:"rateLimit"`
	GasPrice          int64                    `yaml:"gasPrice"`
	GasLimit          int64                    `yaml:"gasLimit"`
	ChainId           int64                    `yaml:"chainId"`
	CompressedService string                   `yaml:"compressedService"`
	DefaultToken      BlockChainTokenConfig    `yaml:"defaultToken"`
	Tokens            []*BlockChainTokenConfig `yaml:"tokens"`
	DefaultNFT        struct {
		Name            string `yaml:"name"`
		ContractAddress string `yaml:"contractAddress"`
		Compressed      bool   `yaml:"compressed"`
		TreePriKey      string `yaml:"tree"`
	} `yaml:"defaultNFT"`
}
