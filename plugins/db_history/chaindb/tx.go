package chaindb

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/KuChainNetwork/kuchain/plugins/db_history/types"
	ptypes "github.com/KuChainNetwork/kuchain/plugins/types"
	"github.com/go-pg/pg/v10"
	"github.com/tendermint/tendermint/libs/log"
)

type txInDB struct {
	ptypes.ReqTx
}
type CreateTxModel struct {
	tableName struct{} `pg:"tx,alias:tx"` // default values are the same

	TxUid      int64  `json:"tx_uid"`
	Height     int64  `json:"height"`
	TxHash     string `json:"tx_hash"`
	Msgs       string `json:"msg"`
	Fee        string `json:"fee"`
	Signatures string `json:"signatures"`
	Memo       string `json:"memo"`
	RawLog     string `json:"raw_log" pg:"raw_log,type:jsonb"`
	To         string `json:"to"`
	From       string `json:"from"`
	Sender     string `json:"sender"`
	Senders    string `json:"senders"`
	Time       string `json:"time"`
}

func newTxInDB(tx ptypes.ReqTx) *txInDB {
	return &txInDB{
		ReqTx: tx,
	}
}

type Signature struct {
	PubKey    string `json:"pub_key"`
	Signature string `json:"signature"`
}

func makeTxmSql(tm ptypes.ReqTx) CreateTxModel {
	bz, _ := json.Marshal(tm.Msgs)
	Msg := string(bz)

	if len(Msg) <= 0 {
		Msg = "{}"
	}

	Hash := strings.ToUpper(hex.EncodeToString(tm.TxHash))
	Fee := tm.Fee.ToString()
	if len(Fee) <= 0 {
		Fee = "{}"
	}

	snowNode, _ := NewSnowNode(0)
	Uid := snowNode.Generate().Int64()

	type signature struct {
		PubKey    string
		Signature []byte
	}
	var tmpSignatures []signature
	for _, p := range tm.Signatures {
		tmpSignatures = append(
			tmpSignatures,
			signature{
				PubKey:    base64.StdEncoding.EncodeToString(p.PubKey.Bytes()),
				Signature: p.Signature,
			},
		)
	}

	bz, _ = json.Marshal(tmpSignatures)
	Sins := string(bz)

	if len(Sins) <= 0 {
		Sins = "{}"
	}

	bz, _ = json.Marshal(tm.RawLog)
	rawLog := string(bz)
	if len(rawLog) <= 0 {
		rawLog = "{}"
	}

	bz, _ = json.Marshal(tm.Senders)
	Sender := string(bz)
	if len(Sender) <= 0 {
		Sender = "{}"
	}

	q := CreateTxModel{
		TxUid:      Uid,
		Height:     tm.Height,
		TxHash:     Hash,
		Msgs:       Msg,
		Fee:        Fee,
		Signatures: Sins,
		Memo:       tm.Memo,
		RawLog:     rawLog,
		Senders:    Sender,
		Time:       tm.Time,
	}

	// todo:临时处理当有多条msgs消息时，取第一条消息来展示，后续需要调整
	if len(tm.Msgs) > 0 {
		type MsgData struct {
			Action string            `json:"action"`
			To     string            `json:"to"`
			Auth   []json.RawMessage `json:"auth"`
			Data   json.RawMessage   `json:"data"`
			From   string            `json:"from"`
			Amount []json.RawMessage `json:"amount"`
			Router string            `json:"router"`
		}

		txMsg := tm.Msgs[0]
		var msg MsgData
		_ = json.Unmarshal(txMsg, &msg)
		if msg.Data == nil || len(msg.Data) <= 0 {
			return q
		}

		if msg.Action == "" && (len(msg.From) != 0 || len(msg.To) != 0) {
			q.From = msg.From
			q.To = msg.To
		}

		requireUnmarshalJSON := func(data []byte, v interface{}) {
			err := json.Unmarshal(data, v)
			if nil != err {
				panic(err)
			}
		}
		type Data struct {
			Type  string          `json:"type"`
			Value json.RawMessage `json:"value"`
		}
		var data Data
		if err := json.Unmarshal(msg.Data, &data); nil != err {
			panic(err)
		}
		switch data.Type {
		case "kuchain/MsgCreateValidator", "kuchain/MsgDelegate", "kuchain/MsgUndelegate":
			type Extra struct {
				ValidatorAccount string `json:"validator_account"`
				DelegatorAccount string `json:"delegator_account"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.DelegatorAccount
			q.To = extra.ValidatorAccount
			q.Sender = extra.DelegatorAccount
		case "account/createData":
			type Extra struct {
				Creator string `json:"creator"`
				Name    string `json:"name"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Creator
			q.To = extra.Name
			q.Sender = extra.Creator
		case "kuchain/MsgBeginRedelegate":
			//staking.MsgBeginRedelegate
			type Extra struct {
				DelegatorAccount    string `json:"delegator_account"`
				ValidatorSrcAccount string `json:"validator_src_account"`
				ValidatorDstAccount string `json:"validator_dst_account"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.DelegatorAccount
			q.To = extra.ValidatorDstAccount
			q.Sender = extra.DelegatorAccount
		case "kuchain/MsgWithdrawDelegationRewardData":
			//distribution.MsgWithdrawDelegatorReward
			type Extra struct {
				ValidatorAddress string `json:"validator_address"`
				DelegatorAddress string `json:"delegator_address"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.DelegatorAddress
			q.To = extra.ValidatorAddress
			q.Sender = extra.DelegatorAddress
		case "account/upAuthData":
			type Extra struct {
				Name string `json:"name"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Name
			//q.To = "account"
			q.Sender = extra.Name
		case "asset/issueData", "asset/createData":
			type Extra struct {
				Creator string `json:"creator"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Creator
			//q.To = "account"
			q.Sender = extra.Creator
		case "kuchain/MsgSubmitProposalBase":
			type Extra struct {
				Proposer string `json:"proposer"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Proposer
			//q.To = "governance"
			q.Sender = extra.Proposer
		case "kuchain/MsgDeposit":
			type Extra struct {
				Depositor string `json:"depositor"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Depositor
			//q.To = "governance"
			q.Sender = extra.Depositor
		case "kuchain/MsgVote":
			type Extra struct {
				Voter string `json:"voter"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Voter
			//q.To = "governance"
			q.Sender = extra.Voter
		case "kuchain/MsgSetWithdrawAccountIdData":
			type Extra struct {
				DelegatorAccountId string `json:"delegator_accountid"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.DelegatorAccountId
			//q.To = "kudistribution"
			q.Sender = extra.DelegatorAccountId
		case "asset/lockData", "asset/unlockData", "asset/burnData", "asset/exerciseData":
			type Extra struct {
				Id string `json:"id"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Id
			//q.To = "asset"
			q.Sender = extra.Id
		case "kuchain/MsgUnjail":
			type Extra struct {
				Address string `json:"address"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.Address
			//q.To = "kuslashing"
			q.Sender = extra.Address
		case "kuchain/MsgEditValidator":
			type Extra struct {
				ValidatorAccount string `json:"validator_account"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			q.From = extra.ValidatorAccount
			//q.To = "kustaking"
			q.Sender = extra.ValidatorAccount
		}
	}

	return q
}

func makeEvent(tm ptypes.ReqTx, logger log.Logger) (Events []types.Event) {
	for _, l := range tm.RawLog.Log {
		for _, e := range l.Events {
			evt := types.Event{
				BlockHeight: tm.Height,
				HashCode:    strings.ToUpper(hex.EncodeToString(tm.TxHash)),
				Type:        e.Type,
			}

			evt.Attributes = make(map[string]string)
			for _, kv := range e.Attributes {
				evt.Attributes[kv.Key] = kv.Value
			}
			Events = append(Events, evt)
		}
	}

	logger.Debug("makeEvent", "evts", Events)
	logger.Debug("makeEvent", "raw", tm.RawLog, "RawCode", tm.RawLog.Code, "height", tm.Height)
	return
}

func InsertTxm(db *pg.DB, logger log.Logger, tx *txInDB) error {
	Events := makeEvent(tx.ReqTx, logger)

	tx_, _ := db.Begin()
	q := makeTxmSql(tx.ReqTx)
	//err := orm.Insert(db, &q)
	_, err := db.Model(&q).Insert()
	if err != nil {
		EventErr(db, logger, NewErrMsg(err))
	}

	InsertTxMsgs(db, logger, tx, tx_, q.TxUid)

	if tx.RawLog.Code == 0 { //fee
		for _, evt := range Events {
			err = InsertEvent(db, logger, &evt)
			if err != nil {
				EventErr(db, logger, NewErrMsg(err))
			}
		}
	}

	tx_.Commit()

	return nil
}
