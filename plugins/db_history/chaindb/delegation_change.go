package chaindb

import (
	"fmt"

	"github.com/KuChainNetwork/kuchain/plugins/db_history/types"
	"github.com/KuChainNetwork/kuchain/utils/eventutil"
	"github.com/go-pg/pg/v10"
	"github.com/tendermint/tendermint/libs/log"
)

type DelegationChange struct {
	Height    string `json:"height"`
	Action    string `json:"action"`
	Hash      string `json:"hash"`
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
	Amount    string `json:"amount"`
	Fee       string `json:"fee"`
	Time      string `json:"time"`
}

type CreateDelegationChangeModel struct {
	tableName struct{} `pg:"delegation_change,alias:delegation_change"` // default values are the same
	ID        int64    // bot

	Height      string `json:"height"`
	Action      string `json:"action"`
	Hash        string `json:"hash"`
	Delegator   string `json:"delegator"`
	Validator   string `json:"validator"`
	Amount      int64  `pg:"default:0" json:"amount"`
	AmountFloat int64  `pg:"default:0" json:"amount_float"`
	Symbol      string `json:"symbol"`
	Fee         string `json:"fee"`
	Time        string `json:"time"`
}

func makeDelegationChangeSql(DeMsg DelegationChange, isAdd bool) CreateDelegationChangeModel {

	coin, _ := NewCoin(DeMsg.Amount)

	q := CreateDelegationChangeModel{
		Height:      DeMsg.Height,
		Action:      DeMsg.Action,
		Hash:        DeMsg.Hash,
		Validator:   DeMsg.Validator,
		Delegator:   DeMsg.Delegator,
		Amount:      coin.Amount,
		AmountFloat: coin.AmountFloat,
		Symbol:      coin.Symbol,
		Fee:         DeMsg.Fee,
		Time:        DeMsg.Time,
	}

	if !isAdd {
		q.Amount = -q.Amount
		q.AmountFloat = -q.AmountFloat
	}

	return q
}

func EventDelegationChange(db *pg.DB, logger log.Logger, evt *types.Event, isAdd bool) {
	var DMsg DelegationChange

	fmt.Println("[INFO]  EventDelegationChange.EventDelegationChange-Attributes", evt.Attributes, "Type", evt.Type)

	err := eventutil.UnmarshalKVMap(evt.Attributes, &DMsg)
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	fmt.Println("[INFO]  EventDelegationChange.EventDelegationChange-UnmarshalAttributes", DMsg.Amount)

	q := makeDelegationChangeSql(DMsg, isAdd)

	fmt.Println("[INFO]  EventDelegationChange.EventDelegationChange-makeChangeSql", q.Amount, q.AmountFloat)

	logger.Debug("EventDelegationChange", "CreateDelegationChangeModel", q)
	//err = orm.Insert(db, &q)
	_, err = db.Model(&q).Insert()
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
	}
}

func EventReDelegate(db *pg.DB, logger log.Logger, evt *types.Event) {
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
	var sourceDelegate CreateDelegationChangeModel
	sourceDelegate.Delegator = attr.Delegator
	sourceDelegate.Validator = attr.SourceValidator
	sourceDelegate.Amount = -amount
	sourceDelegate.AmountFloat = -amountFloat
	sourceDelegate.Symbol = symbol
	sourceDelegate.Time = attr.Time
	sourceDelegate.Height = attr.Height

	_, err = db.Model(&sourceDelegate).Insert()
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	// destination
	var destinationDelegate CreateDelegationChangeModel
	destinationDelegate.Delegator = attr.Delegator
	destinationDelegate.Validator = attr.DestinationValidator
	destinationDelegate.Amount = amount
	destinationDelegate.AmountFloat = amountFloat
	destinationDelegate.Symbol = symbol
	destinationDelegate.Time = attr.Time
	destinationDelegate.Height = attr.Height

	_, err = db.Model(&destinationDelegate).Insert()
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
		return
	}

	fmt.Printf("[INFO] EventReDelegate %+v %+v", sourceDelegate, destinationDelegate)

	_ = tx.Commit()
}
