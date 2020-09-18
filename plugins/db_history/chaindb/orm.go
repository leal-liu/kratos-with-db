package chaindb

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func RegOrm(db *pg.DB, logger log.Logger) error {
	models := []interface{}{
		(*eventInDB)(nil),
		(*txInDB)(nil),
		(*MessageInDB)(nil),
		(*KuTransferInDB)(nil),
		(*blockInDB)(nil),
		(*CreateAccCoinsModel)(nil),
		(*CreateAccountModel)(nil),
		(*BlockInfo)(nil),
		(*CreateCoinTypeModel)(nil),
		(*CreateDelegationModel)(nil),
		(*CreateDelegationChangeModel)(nil),
		(*CreateLockAccCoinsModel)(nil),
		(*EventValidator)(nil),
		(*CreateTxModel)(nil),
		(*CreateTxMsgsModel)(nil),
		(*ErrMsg)(nil),
		(*EventProposerRewardModel)(nil),
		(*EventUnBondModel)(nil),
		(*DelegationReward)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:        false,
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
