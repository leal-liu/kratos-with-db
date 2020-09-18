package chaindb

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/tendermint/tendermint/libs/log"
)

type ErrMsg struct {
	tableName struct{} `pg:"errmsg,alias:errmsg"` // default values are the same
	ID        int64    // both "Id" and "ID" are detected as primary key

	Message string `json:"message"`
	Time    string `json:"time"`
}

func NewErrMsg(err error) ErrMsg {
	e := ErrMsg{
		Message: err.Error(),
		Time:    time.Now().String(),
	}
	return e
}

func NewErrMsgString(str string) ErrMsg {
	e := ErrMsg{
		Message: str,
		Time:    time.Now().String(),
	}
	return e
}

func EventErr(db *pg.DB, logger log.Logger, errIfo ErrMsg) {
	//err := orm.Insert(db, &errIfo)
	_, err := db.Model(&errIfo).Insert()
	if err != nil && logger != nil {
		logger.Error("ErrTableAdd add table error", "err", err.Error())
	}
}
