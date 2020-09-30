package chaindb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KuChainNetwork/kuchain/singleton"
	"github.com/go-pg/pg/v10"
	"github.com/tendermint/tendermint/libs/log"
)

type Msg struct {
	Action string            `json:"action"`
	To     string            `json:"to"`
	Auth   []json.RawMessage `json:"auth"`
	Data   json.RawMessage   `json:"data"`
	From   string            `json:"from"`
	Amount []json.RawMessage `json:"amount"`
	Router string            `json:"router"`
}

type CreateTxMsgsModel struct {
	tableName struct{} `pg:"txmsgs,alias:txmsgs"` // default values are the same
	ID        int64    // both "Id" and "ID" are detected as primary key
	Height    int64    `pg:"default:0" json:"height"`
	TxId      int64    `pg:"btree:t" json:"tx_id"`
	Action    string   `json:"action"`
	To        string   `json:"to"`
	Auth      string   `json:"auth"`
	Data      string   `json:"data"`
	From      string   `json:"from"`
	Amount    string   `json:"amount"`
	Symbol    string   `json:"symbol"`
	Router    string   `json:"router"`
	Sender    string   `json:"sender"`
	Time      string   `json:"time"`
}

func buildTxMsg(logger log.Logger, m json.RawMessage, tx *txInDB, uid int64, sender string) (iMsg CreateTxMsgsModel) {
	var msg Msg
	_ = json.Unmarshal(m, &msg)

	logger.Debug("InsertTxMsgs ", "msg", msg)

	iMsg.TxId = uid
	iMsg.Time = tx.Time
	iMsg.Height = tx.Height
	iMsg.Data = string(msg.Data)
	iMsg.To = msg.To
	iMsg.From = msg.From
	iMsg.Action = msg.Action
	iMsg.Router = msg.Router
	iMsg.Sender = sender

	bz, _ := json.Marshal(msg.Auth)
	iMsg.Auth = string(bz)

	if len(msg.Amount) > 0 {
		for _, ad := range msg.Amount {
			var adn map[string]interface{}
			_ = json.Unmarshal(ad, &adn)

			amount, ok := adn["amount"]
			if ok {
				iMsg.Amount += amount.(string) + " "
			}
			deNo, ok := adn["denom"]
			if ok {
				iMsg.Symbol += deNo.(string) + " "
			}
		}
	} else {
		if msg.Data != nil && len(msg.Data) > 0 {
			type Data struct {
				Value struct {
					Amount []struct {
						Denom  string `json:"denom"`
						Amount string `json:"amount"`
					} `json:"amount"`
				} `json:"value"`
			}
			type Data1 struct {
				Value struct {
					Amount struct {
						Denom  string `json:"denom"`
						Amount string `json:"amount"`
					} `json:"amount"`
				} `json:"value"`
			}

			if strings.Contains(string(msg.Data), "[") {
				var tempMsg Data
				err := json.Unmarshal(msg.Data, &tempMsg)
				if err != nil {
					fmt.Println("[ERROR] txmsg->buildTxMsg()->msg.Data->[]Amount", string(msg.Data))
					panic(err)
				}
				if tempMsg.Value.Amount != nil && len(tempMsg.Value.Amount) > 0 {
					for _, s := range tempMsg.Value.Amount {
						iMsg.Symbol += s.Denom + " "
						iMsg.Amount += s.Amount + " "
					}
				}
			} else {
				var tempMsg Data1
				err := json.Unmarshal(msg.Data, &tempMsg)
				if err != nil {
					fmt.Println("[ERROR] txmsg->buildTxMsg()->msg.Data", string(msg.Data))
					panic(err)
				}

				iMsg.Symbol += tempMsg.Value.Amount.Denom + " "
				iMsg.Amount += tempMsg.Value.Amount.Amount + " "
			}
		}
	}

	if (msg.Data != nil && len(msg.Data) > 0) &&
		(0 >= len(iMsg.From) || 0 >= len(iMsg.To) || 0 >= len(iMsg.Sender)) {
		// process special message
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
			//staking.MsgCreateValidator
			//staking.MsgDelegate
			type Extra struct {
				ValidatorAccount string `json:"validator_account"`
				DelegatorAccount string `json:"delegator_account"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.DelegatorAccount
			iMsg.To = extra.ValidatorAccount
			iMsg.Sender = extra.DelegatorAccount
		case "account/createData":
			type Extra struct {
				Creator string `json:"creator"`
				Name    string `json:"name"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Creator
			iMsg.To = extra.Name
			iMsg.Sender = extra.Creator
		case "kuchain/MsgBeginRedelegate":
			//staking.MsgBeginRedelegate
			type Extra struct {
				DelegatorAccount    string `json:"delegator_account"`
				ValidatorSrcAccount string `json:"validator_src_account"`
				ValidatorDstAccount string `json:"validator_dst_account"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.DelegatorAccount
			iMsg.To = extra.ValidatorDstAccount
			iMsg.Sender = extra.DelegatorAccount
		case "kuchain/MsgWithdrawDelegationRewardData":
			//distribution.MsgWithdrawDelegatorReward
			type Extra struct {
				ValidatorAddress string `json:"validator_address"`
				DelegatorAddress string `json:"delegator_address"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.DelegatorAddress
			iMsg.To = extra.ValidatorAddress
			iMsg.Sender = extra.DelegatorAddress
		case "account/upAuthData":
			type Extra struct {
				Name string `json:"name"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Name
			//iMsg.To = "account"
			iMsg.Sender = extra.Name
		case "asset/issueData":
			type Extra struct {
				Creator string `json:"creator"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Creator
			//iMsg.To = "account"
			iMsg.Sender = extra.Creator
		case "asset/createData":
			type Extra struct {
				Creator   string `json:"creator"`
				MaxSupply struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				} `json:"max_supply"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Creator
			//iMsg.To = "account"
			iMsg.Sender = extra.Creator
			iMsg.Symbol = extra.MaxSupply.Denom
		case "kuchain/MsgSubmitProposalBase":
			type Extra struct {
				Proposer string `json:"proposer"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Proposer
			//iMsg.To = "governance"
			iMsg.Sender = extra.Proposer
		case "kuchain/MsgDeposit":
			type Extra struct {
				Depositor string `json:"depositor"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Depositor
			//iMsg.To = "governance"
			iMsg.Sender = extra.Depositor
		case "kuchain/MsgVote":
			type Extra struct {
				Voter string `json:"voter"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Voter
			//iMsg.To = "governance"
			iMsg.Sender = extra.Voter
		case "kuchain/MsgSetWithdrawAccountIdData":
			type Extra struct {
				DelegatorAccountId string `json:"delegator_accountid"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.DelegatorAccountId
			//iMsg.To = "kudistribution"
			iMsg.Sender = extra.DelegatorAccountId
		case "asset/lockData", "asset/unlockData", "asset/burnData", "asset/exerciseData":
			type Extra struct {
				Id string `json:"id"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Id
			//iMsg.To = "asset"
			iMsg.Sender = extra.Id
		case "kuchain/MsgUnjail":
			type Extra struct {
				Address string `json:"address"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.Address
			//iMsg.To = "kuslashing"
			iMsg.Sender = extra.Address
		case "kuchain/MsgEditValidator":
			type Extra struct {
				ValidatorAccount string `json:"validator_account"`
			}
			var extra Extra
			requireUnmarshalJSON(data.Value, &extra)
			iMsg.From = extra.ValidatorAccount
			//iMsg.To = "kustaking"
			iMsg.Sender = extra.ValidatorAccount
		}
	}

	if len(strings.Trim(iMsg.Symbol, " ")) == 0 {
		iMsg.Symbol = singleton.MainCoinsSymbol
	}

	logger.Debug("buildTxMsg", "iMsg", iMsg)

	return
}

func InsertTxMsgs(db *pg.DB, logger log.Logger, tx *txInDB, _ *pg.Tx, uid int64) bool {
	for _, m := range tx.Msgs {
		iMsg := buildTxMsg(logger, m, tx, uid, "")
		//err := orm.Insert(db, &iMsg)
		_, err := db.Model(&iMsg).Insert()
		if err != nil {
			EventErr(db, logger, NewErrMsg(err))
		}
	}
	return true
}
