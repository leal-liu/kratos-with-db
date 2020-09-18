package types

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
)

type ReqBlock struct {
	abci.RequestBeginBlock
	ValidatorInfo string
	Time          time.Time
}
