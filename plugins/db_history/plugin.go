package dbHistory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/KuChainNetwork/kuchain/cache"
	"github.com/KuChainNetwork/kuchain/plugins/db_history/chaindb"
	"github.com/KuChainNetwork/kuchain/plugins/db_history/config"
	"github.com/KuChainNetwork/kuchain/plugins/db_history/types"
	types2 "github.com/KuChainNetwork/kuchain/plugins/types"
	"github.com/KuChainNetwork/kuchain/singleton"
	"github.com/KuChainNetwork/kuchain/synchronizer"
	"github.com/go-pg/pg/v10/orm"
	"github.com/go-redis/redis/v8"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	maxAssetSyncConcurrency        = 256
	maxDistributionSyncConcurrency = 256
	syncTimeout                    = 10 * time.Second
)

// plugin for test
type plugin struct {
	logger log.Logger

	cfg config.Cfg
	db  *dbService
	wg  sync.WaitGroup
}

func (t *plugin) Init(ctx types.Context) error {
	t.logger.Info("plugin init", "name", types.PluginName)
	t.db = NewDB(t.cfg, ctx.Logger().With("module", "his-database"))

	if chaindb.ErrDatabase == nil {
		chaindb.ErrDatabase = t.db.errDatabase
	}

	t.logger.Info("plugin init", "name", types.PluginName)
	t.db = NewDB(t.cfg, ctx.Logger().With("module", "his-database"))
	if chaindb.ErrDatabase == nil {
		chaindb.ErrDatabase = t.db.errDatabase
	}

	// startup sync services
	if t.cfg.AccountCoinsSync {
		t.startAssetSyncService()
	}
	if t.cfg.DistributionRewardSync {
		t.startDelegationRewardSync()
	}

	// init Redis
	singleton.Redis = redis.NewClient(&redis.Options{
		Addr:     t.cfg.Redis.Address,
		Password: t.cfg.Redis.Password,
		DB:       t.cfg.Redis.Db,
	})
	_, err := singleton.Redis.Ping(context.Background()).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	// init MainCoinsSymbol
	singleton.MainCoinsSymbol = t.cfg.MainCoinsSymbol

	return nil
}

func (t *plugin) Start(ctx types.Context) error {
	t.logger.Info("plugin start", "name", types.PluginName)

	if err := t.db.Start(); err != nil {
		return err
	}

	return nil
}

func (t *plugin) Stop(ctx types.Context) error {
	t.logger.Info("plugin stop", "name", types.PluginName)

	if err := t.db.Stop(); err != nil {
		return err
	}

	return nil
}

func (t *plugin) MsgHandler() types.PluginMsgHandler {
	return func(ctx types.Context, msg sdk.Msg) {
		t.OnMsg(ctx, msg)
	}
}

func (t *plugin) TxHandler() types.PluginTxHandler {
	return func(ctx types.Context, tx types2.ReqTx) {
		t.OnTx(ctx, tx)
	}
}

func (t *plugin) EvtHandler() types.PluginEvtHandler {
	return func(ctx types.Context, evt types.Event) {
		t.OnEvent(ctx, evt)
	}
}

func (t *plugin) Logger() log.Logger {
	return t.logger
}

func (t *plugin) Name() string {
	return types.PluginName
}

// New new plugin
func New(ctx types.Context, cfg types.BaseCfg) *plugin {
	logger := ctx.Logger().With("module", fmt.Sprintf("plugins/%s", types.PluginName))

	res := &plugin{
		logger: logger,
	}

	if err := json.Unmarshal(cfg.CfgRaw, &res.cfg); err != nil {
		panic(err)
	}

	logger.Info("new plugin", "name", types.PluginName, "cfg", res.cfg)

	return res
}

func (t *plugin) convertCoin(amountStr string) (amount int64, amountFloat int64, err error) {
	if len(amountStr) <= 0 {
		err = fmt.Errorf("amountStr length is 0")
		return
	}

	for i, v := range amountStr {
		if v < '0' || v > '9' {
			amountStr = amountStr[:i]
			break
		}
	}
	fmt.Println("------------->convertCoin", amountStr)
	if len(amountStr) <= 18 {
		amount = 0
		amountFloat, err = strconv.ParseInt(amountStr, 10, 64)
		return
	}

	amount, err = strconv.ParseInt(amountStr[:len(amountStr)-18], 10, 64)
	if err != nil {
		return
	}
	amountFloat, err = strconv.ParseInt(amountStr[len(amountStr)-18:], 10, 64)
	if err != nil {
		return
	}
	return
}

// startAssetSyncService
func (t *plugin) startAssetSyncService() {
	if _, err := orm.NewQuery(t.db.database, &chaindb.CreateAccCoinsModel{}).
		Set("sync_state=?", 0).
		Where("true").
		Update(); nil != err {
		panic(err)
	}
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		duration := 1 * time.Second
		tick := time.NewTimer(duration)
		for {
			select {
			case <-tick.C:
				var list []*chaindb.CreateAccCoinsModel
				err := orm.NewQuery(t.db.database, &list).Where("sync_state=?", 0).Select()
				if nil != err {
					panic(err)
				}
				for 0 < len(list) {
					chunkSize := maxAssetSyncConcurrency
					if chunkSize > len(list) {
						chunkSize = len(list)
					}
					var wg sync.WaitGroup
					wg.Add(chunkSize)
					for i := 0; i < chunkSize; i++ {
						m := list[i]
						m.SyncState = 1
						if _, err = orm.NewQuery(t.db.database, m).
							Set("sync_state=?", m.SyncState).
							Where("id=? and sync_state=?", m.ID, 0).Update(); nil != err {
							panic(err)
						}
						go func(m *chaindb.CreateAccCoinsModel) {
							defer wg.Done()
							syncTool := synchronizer.NewAssetSync(t.cfg.Chain.Host, t.cfg.Chain.Port)
							coins, err := syncTool.Sync(m.Account, m.Symbol, syncTimeout)
							if nil == err {
								if nil != coins && 0 < len(coins) {
									amount, amountFloat, err := t.convertCoin(coins[0])
									//coin, err := chaindb.NewCoin(coins[0])
									if nil != err {
										panic(err)
									}
									m.Amount = amount
									m.AmountFloat = amountFloat
									m.SyncState = 2
								} else {
									m.SyncState = 0
								}
								if _, err = orm.NewQuery(t.db.database, m).
									Set("sync_state=?, amount=?, amount_float=?", m.SyncState, m.Amount, m.AmountFloat).
									Where("id=? and sync_state=?", m.ID, 1).
									Update(); nil != err {
									panic(err)
								}
								key := fmt.Sprintf("account_coins:%s", m.Account)
								err = cache.NewRedisClientWrapper(singleton.Redis).
									DelAccountCoinsModel(key, 3*time.Second)
							} else {
								t.logger.Error(fmt.Sprintf("account: %s, symbol: %s, sync asset error: %s\n",
									m.Account,
									m.Symbol,
									err))
								if _, err = orm.NewQuery(t.db.database, m).
									Set("sync_state=?", 0).
									Where("id=? and sync_state=?", m.ID, 1).
									Update(); nil != err {
									panic(err)
								}
							}
						}(m)
					}
					wg.Wait()
					list = list[chunkSize:]
				}
				tick.Reset(duration)
			}
		}
	}()
}

func (t *plugin) startDelegationRewardSync() {
	if _, err := orm.NewQuery(t.db.database, &chaindb.DelegationReward{}).
		Set("sync_state=?", 0).
		Where("true").
		Update(); nil != err {
		panic(err)
	}
	t.wg.Add(1)
	go func() {
		defer func() {
			t.wg.Done()
		}()

		duration := 1 * time.Second
		tick := time.NewTimer(duration)
		for {
			select {
			case <-tick.C:
				var list []*chaindb.DelegationReward
				err := orm.NewQuery(t.db.database, &list).Where("sync_state=?", 0).Select()
				if nil != err {
					panic(err)
				}
				for 0 < len(list) {
					chunkSize := maxDistributionSyncConcurrency
					if chunkSize > len(list) {
						chunkSize = len(list)
					}
					var wg sync.WaitGroup
					wg.Add(chunkSize)
					for i := 0; i < chunkSize; i++ {
						m := list[i]
						m.SyncState = 1
						var result orm.Result
						if result, err = orm.NewQuery(t.db.database, m).
							Set("sync_state=?", m.SyncState).
							Where("validator=? and delegator=?", m.Validator, m.Delegator).
							Where("sync_state=?", 0).Update(); nil != err {
							panic(err)
						}
						if 0 >= result.RowsAffected() {
							wg.Done()
							continue
						}
						go func(m *chaindb.DelegationReward) {
							defer wg.Done()
							syncTool := synchronizer.NewDistributionSync(t.cfg.Chain.Host, t.cfg.Chain.Port)
							coins, err := syncTool.Sync(m.Validator, m.Delegator, syncTimeout)
							if nil == err {
								if 0 < len(coins) {
									m.Amount = singleton.CdcInst.MustMarshalJSON(coins)
									m.SyncState = 2
								} else {
									m.SyncState = 0
								}
								if m.Amount == nil || len(m.Amount) == 0 {
									m.Amount = []byte("[]")
								}
								if _, err = orm.NewQuery(t.db.database, m).
									Set("sync_state=?", m.SyncState).
									Set("amount=?", m.Amount).
									Where("sync_state=? and id=?", 1, m.ID).Update(); nil != err {
									fmt.Println("[ERROR] sync->Amount", m.Amount, "err", err)
									panic(err)
								}
								key := fmt.Sprintf("%s:%s", "delegation_reward", m.Validator)
								err = cache.NewRedisClientWrapper(singleton.Redis).
									DelDelegationRewardModel(key, 3*time.Second)
							} else {
								t.logger.Error(fmt.Sprintf("validator: %s, delegator: %s, sync delegation reward error: %s",
									m.Validator,
									m.Delegator,
									err))
								_, _ = fmt.Fprintln(os.Stderr, err)
								if _, err = orm.NewQuery(t.db.database, m).
									Set("sync_state=?", 0).
									Where("sync_state=? and id=?", 1, m.ID).Update(); nil != err {
									panic(err)
								}
							}
						}(m)
					}
					wg.Wait()
					list = list[chunkSize:]
				}
				tick.Reset(duration)
			}
		}
	}()
}
