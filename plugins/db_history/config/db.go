package config

// DBCfg cfg for database connect
type DBCfg struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// ChainCfg cfg for chain connect
type ChainCfg struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type RedisCfg struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

type Cfg struct {
	DB                     DBCfg    `json:"db"`
	Chain                  ChainCfg `json:"chain"`
	AccountCoinsSync       bool     `json:"account_coins_sync"`
	DistributionRewardSync bool     `json:"distribution_reward_sync"`
	Redis                  RedisCfg `json:"redis"`
	MainCoinsSymbol        string   `json:"main_coins_symbol"`
}
