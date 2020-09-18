package chaindb

import (
	"encoding/json"
	"fmt"

	"github.com/KuChainNetwork/kuchain/plugins/db_history/types"
	"github.com/KuChainNetwork/kuchain/utils/eventutil"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type DelegationReward struct {
	tableName struct{}        `pg:"delegationreward,alias:delegationreward"` // default values are the same
	ID        int             // both "Id" and "ID" are detected as primary key
	Height    string          `pg:"default:0" json:"height"`
	Validator string          `json:"validator"`
	Delegator string          `json:"delegator"`
	Amount    json.RawMessage `pg:"amount;type:json" json:"amount"`
	SyncState int             `pg:"default:0" json:"sync_state"`
	Time      string          `json:"time"`
}

type DelegationRewardService struct {
	db *pg.DB
}

func (object *DelegationRewardService) AddEvent(e *types.Event) (err error) {
	var msg Delegation
	if err = eventutil.UnmarshalKVMap(e.Attributes, &msg); err != nil {
		return
	}
	err = object.AddModel(&DelegationReward{
		Height:    msg.Height,
		Validator: msg.Validator,
		Delegator: msg.Delegator,
		Time:      msg.Time,
	})

	if err != nil {
		fmt.Println("[ERROR]: DelegationRewardService.AddEvent", err)
	}

	return
}

func (object *DelegationRewardService) AddModel(model *DelegationReward) (err error) {
	var count int
	if count, err = orm.NewQuery(object.db, model).Where("validator=? and delegator=?",
		model.Validator,
		model.Delegator).Count(); nil != err {
		return
	}
	if 0 >= count {
		//err = orm.Insert(object.db, model)
		_, err = object.db.Model(model).Insert()
	}
	return
}

func (object *DelegationRewardService) AddEventByReDelegator(e *types.Event) (err error) {
	type attributes struct {
		SourceValidator      string `json:"source_validator"`
		DestinationValidator string `json:"destination_validator"`
		Amount               string `json:"amount"`
		Delegator            string `json:"delegator"`
		Time                 string `json:"block_time"`
		Height               string `json:"height"`
	}

	var msg attributes

	if err = eventutil.UnmarshalKVMap(e.Attributes, &msg); err != nil {
		return
	}
	err = object.AddModel(&DelegationReward{
		Height:    msg.Height,
		Validator: msg.DestinationValidator,
		Delegator: msg.Delegator,
		Time:      msg.Time,
	})

	if err != nil {
		fmt.Println("[ERROR]: DelegationRewardService.AddEvent", err)
	}

	return
}

func NewDelegationRewardService(db *pg.DB) *DelegationRewardService {
	return &DelegationRewardService{db: db}
}
