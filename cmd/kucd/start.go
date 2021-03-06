package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/KuChainNetwork/kuchain/plugins"
	"github.com/KuChainNetwork/kuchain/singleton"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abciServer "github.com/tendermint/tendermint/abci/server"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	dbm "github.com/tendermint/tm-db"
)

// Tendermint full-node start flags
const (
	flagWithTendermint       = "with-tendermint"
	flagAddress              = "address"
	flagTraceStore           = "trace-store"
	flagPruning              = "pruning"
	flagPruningKeepEvery     = "pruning-keep-every"
	flagPruningSnapshotEvery = "pruning-snapshot-every"
	flagCPUProfile           = "cpu-profile"
	FlagMinGasPrices         = "minimum-gas-prices"
	FlagHaltHeight           = "halt-height"
	FlagHaltTime             = "halt-time"
	FlagInterBlockCache      = "inter-block-cache"
	FlagUnsafeSkipUpgrades   = "unsafe-skip-upgrades"
	FlagPluginCfgPath        = "plugin-cfg"
)

var (
	errPruningWithGranularOptions = fmt.Errorf(
		"'--%s' flag is not compatible with granular options  '--%s' or '--%s'",
		flagPruning, flagPruningKeepEvery, flagPruningSnapshotEvery,
	)
	errPruningGranularOptions = fmt.Errorf(
		"'--%s' and '--%s' must be set together",
		flagPruningSnapshotEvery, flagPruningKeepEvery,
	)
)

// StartCmd runs the service passed in, either stand-alone or in-process with
// Tendermint.
func StartCmd(ctx *server.Context, appCreator server.AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with Tendermint in or out of process. By
default, the application will run with Tendermint in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-snapshot-every' and 'pruning-keep-every' together.

For '--pruning' the options are as follows:

syncable: only those states not needed for state syncing will be deleted (flushes every 100th to disk and keeps every 10000th)
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current state

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkPruningParams()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !viper.GetBool(flagWithTendermint) {
				ctx.Logger.Info("starting ABCI without Tendermint")
				return startStandAlone(ctx, appCreator)
			}

			ctx.Logger.Info("starting ABCI with Tendermint")

			_, err := startInProcess(ctx, appCreator)
			return err
		},
	}

	// core flags for the ABCI application
	cmd.Flags().Bool(flagWithTendermint, true, "Run abci app embedded in-process with tendermint")
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(flagPruning, "syncable", "Pruning strategy: syncable, nothing, everything")
	cmd.Flags().Int64(flagPruningKeepEvery, 0, "Define the state number that will be kept")
	cmd.Flags().Int64(flagPruningSnapshotEvery, 0, "Defines the state that will be snapshot for pruning")
	cmd.Flags().String(
		FlagMinGasPrices, "",
		"Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)",
	)
	cmd.Flags().IntSlice(FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().String(FlagPluginCfgPath, "", "Config file path for plugins")

	// add support for all Tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

// checkPruningParams checks that the provided pruning params are correct
func checkPruningParams() error {
	if !viper.IsSet(flagPruning) && !viper.IsSet(flagPruningKeepEvery) && !viper.IsSet(flagPruningSnapshotEvery) {
		return nil
	}

	if viper.IsSet(flagPruning) {
		if viper.IsSet(flagPruningKeepEvery) || viper.IsSet(flagPruningSnapshotEvery) {
			return errPruningWithGranularOptions
		}

		return nil
	}

	if !(viper.IsSet(flagPruningKeepEvery) && viper.IsSet(flagPruningSnapshotEvery)) {
		return errPruningGranularOptions
	}

	return nil
}

func startStandAlone(ctx *server.Context, appCreator server.AppCreator) error {
	addr := viper.GetString(flagAddress)
	home := viper.GetString("home")
	traceWriterFile := viper.GetString(flagTraceStore)

	db, err := openDB(home)
	if err != nil {
		return err
	}
	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		return err
	}

	app := appCreator(ctx.Logger, db, traceWriter)

	svr, err := abciServer.NewServer(addr, "socket", app)
	if err != nil {
		return fmt.Errorf("error creating listener: %v", err)
	}

	svr.SetLogger(ctx.Logger.With("module", "abci-server"))

	err = svr.Start()
	if err != nil {
		tmos.Exit(err.Error())
	}

	tmos.TrapSignal(ctx.Logger, func() {
		// cleanup
		err = svr.Stop()
		if err != nil {
			tmos.Exit(err.Error())
		}
	})

	// run forever (the node will not be returned)
	select {}
}

func initPlugins(ctx *server.Context) error {
	cfgFilePath := viper.GetString(FlagPluginCfgPath)
	if cfgFilePath == "" {
		ctx.Logger.Debug("no need start plugins")
		return nil
	}

	pluginCfg := struct {
		Plugins []plugins.BaseCfg
	}{}

	raws, err := tmos.ReadFile(cfgFilePath)
	if err != nil {
		return errors.Wrapf(err, "read plugin file %s err", cfgFilePath)
	}

	if err := json.Unmarshal(raws, &pluginCfg); err != nil {
		return errors.Wrapf(err, "unmarshal plugin config")
	}

	pluginCtx := plugins.NewContext(ctx.Logger)
	return plugins.InitPlugins(pluginCtx, pluginCfg.Plugins)
}

func startInProcess(ctx *server.Context, appCreator server.AppCreator) (*node.Node, error) {
	cfg := ctx.Config
	home := cfg.RootDir

	if err := initPlugins(ctx); err != nil {
		tmos.Exit(err.Error())
	}

	traceWriterFile := viper.GetString(flagTraceStore)
	db, err := openDB(home)
	if err != nil {
		return nil, err
	}

	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		return nil, err
	}

	app := appCreator(ctx.Logger, db, traceWriter)

	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	// create & start tendermint node
	tmNode, err := node.NewNode(
		cfg,
		pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation),
		ctx.Logger.With("module", "node"),
	)
	if err != nil {
		return nil, err
	}

	singleton.NodeInst = tmNode

	if err := tmNode.Start(); err != nil {
		return nil, err
	}

	var cpuProfileCleanup func()

	if cpuProfile := viper.GetString(flagCPUProfile); cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return nil, err
		}

		ctx.Logger.Info("starting CPU profiler", "profile", cpuProfile)
		if err := pprof.StartCPUProfile(f); err != nil {
			return nil, err
		}

		cpuProfileCleanup = func() {
			ctx.Logger.Info("stopping CPU profiler", "profile", cpuProfile)
			pprof.StopCPUProfile()
			f.Close()
		}
	}

	server.TrapSignal(func() {
		if tmNode.IsRunning() {
			_ = tmNode.Stop()
		}

		if cpuProfileCleanup != nil {
			cpuProfileCleanup()
		}

		plugins.StopPlugins(plugins.NewContext(ctx.Logger))

		ctx.Logger.Info("exiting...")

		loggerSync, ok := ctx.Logger.(interface {
			Flush() error
		})
		if ok {
			loggerSync.Flush()
		}
	})

	// run forever (the node will not be returned)
	select {}
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	db, err := sdk.NewLevelDB("application", dataDir)
	return db, err
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile != "" {
		w, err = os.OpenFile(
			traceWriterFile,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0666,
		)
		return
	}
	return
}
