package chaindb

import (
	"fmt"

	"github.com/KuChainNetwork/kuchain/plugins/db_history/types"
	"github.com/KuChainNetwork/kuchain/utils/eventutil"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/tendermint/tendermint/libs/log"
)

type Delegation struct {
	Height    string `json:"height"`
	Validator string `json:"validator"`
	Delegator string `json:"delegator"`
	Amount    string `json:"amount"`
	Time      string `json:"block_time"`
}

type CreateDelegationModel struct {
	tableName struct{} `pg:"delegation,alias:delegation"` // default values are the same

	ID          int    // both "Id" and "ID" are detected as primary key
	Height      string `pg:"default:0" json:"height"`
	Validator   string `pg:"unique:vd" json:"validator"`
	Delegator   string `pg:"unique:vd" json:"delegator"`
	Amount      int64  `pg:"default:0" json:"amount"`
	AmountFloat int64  `pg:"default:0" json:"amount_float"`
	Symbol      string `json:"symbol"`
	Time        string `json:"time"`
}

func makeDelegationSql(msg Delegation, isAdd bool) CreateDelegationModel {
	coin, _ := NewCoin(msg.Amount)

	q := CreateDelegationModel{
		Height:      msg.Height,
		Validator:   msg.Validator,
		Delegator:   msg.Delegator,
		Amount:      coin.Amount,
		AmountFloat: coin.AmountFloat,
		Symbol:      coin.Symbol,
		Time:        msg.Time,
	}

	if !isAdd {
		q.Amount = -q.Amount
		q.AmountFloat = -q.AmountFloat
	}

	return q
}

func dExec(db *pg.DB, model CreateDelegationModel, logger log.Logger) error {
	var m CreateDelegationModel
	err := orm.NewQuery(db, &m).Where(fmt.Sprintf("validator='%s' and delegator='%s'", model.Validator, model.Delegator)).Select()
	if err != nil {
		logger.Debug("dExec1", "model", model)
		//err = orm.Insert(db, &model)
		_, err = db.Model(&model).Insert()
	} else {
		model.Amount, model.AmountFloat = CoinAdd(model.Amount, model.AmountFloat, m.Amount, m.AmountFloat)
		logger.Debug("dExec2", "model", model)
		_, err = orm.NewQuery(db, &model).Where(fmt.Sprintf("validator='%s' and delegator='%s'", model.Validator, model.Delegator)).Update()
	}

	if err == nil {
		_, err = orm.NewQuery(db, &model).
			Where(fmt.Sprintf("validator='%s' and delegator='%s'", model.Validator, model.Delegator)).
			Set(fmt.Sprintf("amount=%d, amount_float=%d", model.Amount, model.AmountFloat)).Update()
	}
	return err
}

func EventDelegationAdd(db *pg.DB, logger log.Logger, evt *types.Event) {
	var msg Delegation
	err := eventutil.UnmarshalKVMap(evt.Attributes, &msg)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	q := makeDelegationSql(msg, true)
	err = dExec(db, q, logger)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
	}
}

func EventDelegationDel(db *pg.DB, logger log.Logger, evt *types.Event) {
	var msg Delegation
	err := eventutil.UnmarshalKVMap(evt.Attributes, &msg)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	q := makeDelegationSql(msg, false)
	err = dExec(db, q, logger)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
	}
}

func EventReDelegationChange(db *pg.DB, logger log.Logger, evt *types.Event) {
	tx, _ := db.Begin()

	type attributes struct {
		SourceValidator      string `json:"source_validator"`
		DestinationValidator string `json:"destination_validator"`
		Amount               string `json:"amount"`
		Delegator            string `json:"delegator"`
		Time                 string `json:"block_time"`
		Height               string `json:"height"`
	}

	var attr attributes

	err := eventutil.UnmarshalKVMap(evt.Attributes, &attr)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	amountStr, symbol, err := splitSymbol(attr.Amount)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	amount, amountFloat, err := parseAmountStr(amountStr)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	// source
	var sourceDelegate CreateDelegationModel
	sourceDelegate.Delegator = attr.Delegator
	sourceDelegate.Validator = attr.SourceValidator
	sourceDelegate.Amount = -amount
	sourceDelegate.AmountFloat = -amountFloat
	sourceDelegate.Symbol = symbol
	sourceDelegate.Time = attr.Time
	sourceDelegate.Height = attr.Height

	err = dExec(db, sourceDelegate, logger)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	// destination
	var destinationDelegate CreateDelegationModel
	destinationDelegate.Delegator = attr.Delegator
	destinationDelegate.Validator = attr.DestinationValidator
	destinationDelegate.Amount = amount
	destinationDelegate.AmountFloat = amountFloat
	destinationDelegate.Symbol = symbol
	destinationDelegate.Time = attr.Time
	destinationDelegate.Height = attr.Height

	err = dExec(db, destinationDelegate, logger)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	_ = tx.Commit()
}
