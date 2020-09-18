package singleton

import (
	"github.com/go-redis/redis/v8"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/node"
)

var (
	CdcInst         *amino.Codec
	NodeInst        *node.Node
	Redis           *redis.Client
	MainCoinsSymbol string
)
