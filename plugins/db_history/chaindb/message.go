package chaindb

import (
	chainTypes "github.com/KuChainNetwork/kuchain/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

type MessageInDB struct {
	tableName struct{} `pg:"messages,alias:messages"` // default values are the same

	ID int64 // both "Id" and "ID" are detected as primary key

	sdk.Msg // FIXME: use petty show msg
}

type KuTransferInDB struct {
	tableName struct{} `pg:"transfer,alias:transfer"` // default values are the same

	ID     int64 // both "Id" and "ID" are detected as primary key
	Route  string
	Type   string
	From   string
	To     string
	Amount int64
	Symbol string
}

func newMsgToDB(msg sdk.Msg) *MessageInDB {
	return &MessageInDB{
		Msg: msg,
	}
}

func processMsg(db *pg.DB, msg sdk.Msg) error {
	if _, err := db.Model(newMsgToDB(msg)).Insert(); err != nil {
		return errors.Wrapf(err, "insert msg")
	}

	if msg, ok := msg.(chainTypes.KuTransfMsg); ok {
		amounts := msg.GetAmount()

		in := &KuTransferInDB{
			Route: msg.Route(),
			Type:  msg.Type(),
			From:  msg.GetFrom().String(),
			To:    msg.GetTo().String(),
		}

		for _, amount := range amounts {
			in.Amount = amount.Amount.BigInt().Int64()
			in.Symbol = amount.Denom
			if _, err := db.Model(in).Insert(); err != nil {
				if err != nil {
					EventErr(db, nil, NewErrMsg(err))
				}
				return err
			}
		}
	}

	return nil
}
