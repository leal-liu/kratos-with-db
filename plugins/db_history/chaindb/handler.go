package chaindb

import (
	"reflect"

	"github.com/KuChainNetwork/kuchain/plugins/test/types"
	types2 "github.com/KuChainNetwork/kuchain/plugins/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-pg/pg/v10"
	"github.com/tendermint/tendermint/libs/log"
)

func Process(db *pg.DB, logger log.Logger, msg interface{}) error {
	logger.Debug("process msg", "typ", reflect.TypeOf(msg), "msg", msg)
	switch msg := msg.(type) {
	case types.Event:
		return InsertEvent(db, logger, &msg)
	case types2.ReqTx:
		return InsertTxm(db, logger, newTxInDB(msg))
	case types2.ReqBlock:
		return InsertBlockInfo(db, logger, newBlockInDB(msg))
	}

	if msg, ok := msg.(sdk.Msg); ok {
		return processMsg(db, msg)
	}

	return nil
}

func insert(db *pg.DB, obj interface{}) error {
	_, err := db.Model(&obj).Insert()

	//return db.Insert(obj)
	return err
}
