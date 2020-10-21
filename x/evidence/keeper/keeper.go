package keeper

import (
	"fmt"

	chaintype "github.com/KuChainNetwork/kuchain/chain/types"

	"github.com/KuChainNetwork/kuchain/x/evidence/exported"
	"github.com/KuChainNetwork/kuchain/x/evidence/external"
	"github.com/KuChainNetwork/kuchain/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper defines the evidence module's keeper. The keeper is responsible for
// managing persistence, state transitions and query handling for the evidence
// module.
type Keeper struct {
	storeKey       sdk.StoreKey
	paramSpace     external.ParamsSubspace
	router         types.Router
	stakingKeeper  types.StakingKeeper
	slashingKeeper types.SlashingKeeper
}

func NewKeeper(
	storeKey sdk.StoreKey, paramSpace external.ParamsSubspace,
	stakingKeeper types.StakingKeeper, slashingKeeper types.SlashingKeeper,
) *Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:       storeKey,
		paramSpace:     paramSpace,
		stakingKeeper:  stakingKeeper,
		slashingKeeper: slashingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetRouter sets the Evidence Handler router for the x/evidence module. Note,
// we allow the ability to set the router after the Keeper is constructed as a
// given Handler may need access the Keeper before being constructed. The router
// may only be set once and will be sealed if it's not already sealed.
func (k *Keeper) SetRouter(rtr types.Router) {
	// It is vital to seal the Evidence Handler router as to not allow further
	// handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	if !rtr.Sealed() {
		rtr.Seal()
	}
	if k.router != nil {
		panic(fmt.Sprintf("attempting to reset router on x/%s", types.ModuleName))
	}

	k.router = rtr
}

// GetEvidenceHandler returns a registered Handler for a given Evidence type. If
// no handler exists, an error is returned.
func (k Keeper) GetEvidenceHandler(evidenceRoute string) (types.Handler, error) {
	if !k.router.HasRoute(evidenceRoute) {
		return nil, sdkerrors.Wrap(types.ErrNoEvidenceHandlerExists, evidenceRoute)
	}

	return k.router.GetRoute(evidenceRoute), nil
}

// SubmitEvidence attempts to match evidence against the keepers router and execute
// the corresponding registered Evidence Handler. An error is returned if no
// registered Handler exists or if the Handler fails. Otherwise, the evidence is
// persisted.
func (k Keeper) SubmitEvidence(ctx sdk.Context, evidence exported.Evidence) error {
	if _, ok := k.GetEvidence(ctx, evidence.Hash()); ok {
		return sdkerrors.Wrap(types.ErrEvidenceExists, evidence.Hash().String())
	}
	if !k.router.HasRoute(evidence.Route()) {
		return sdkerrors.Wrap(types.ErrNoEvidenceHandlerExists, evidence.Route())
	}

	handler := k.router.GetRoute(evidence.Route())
	if err := handler(ctx, evidence); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidEvidence, err.Error())
	}

	ctx.EventManager().EmitEvent(
		chaintype.NewEvent(ctx,
			types.EventTypeSubmitEvidence,
			sdk.NewAttribute(types.AttributeKeyEvidenceHash, evidence.Hash().String()),
		),
	)

	k.SetEvidence(ctx, evidence)
	return nil
}

// SetEvidence sets Evidence by hash in the module's KVStore.
func (k Keeper) SetEvidence(ctx sdk.Context, evidence exported.Evidence) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEvidence)
	store.Set(evidence.Hash(), k.MustMarshalEvidence(evidence))
}

// GetEvidence retrieves Evidence by hash if it exists. If no Evidence exists for
// the given hash, (nil, false) is returned.
func (k Keeper) GetEvidence(ctx sdk.Context, hash tmbytes.HexBytes) (exported.Evidence, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEvidence)

	bz := store.Get(hash)
	if len(bz) == 0 {
		return nil, false
	}

	return k.MustUnmarshalEvidence(bz), true
}

// IterateEvidence provides an interator over all stored Evidence objects. For
// each Evidence object, cb will be called. If the cb returns true, the iterator
// will close and stop.
func (k Keeper) IterateEvidence(ctx sdk.Context, cb func(exported.Evidence) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEvidence)
	iterator := sdk.KVStorePrefixIterator(store, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		evidence := k.MustUnmarshalEvidence(iterator.Value())

		if cb(evidence) {
			break
		}
	}
}

// GetAllEvidence returns all stored Evidence objects.
func (k Keeper) GetAllEvidence(ctx sdk.Context) (evidence []exported.Evidence) {
	k.IterateEvidence(ctx, func(e exported.Evidence) bool {
		evidence = append(evidence, e)
		return false
	})

	return evidence
}

// MustUnmarshalEvidence attempts to decode and return an Evidence object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalEvidence(bz []byte) exported.Evidence {
	evidence, err := types.Evidence_Cdc.UnmarshalEvidence(bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode evidence: %w", err))
	}

	return evidence
}

// MustMarshalEvidence attempts to encode an Evidence object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalEvidence(evidence exported.Evidence) []byte {
	bz, err := types.Evidence_Cdc.MarshalEvidence(evidence)
	if err != nil {
		panic(fmt.Errorf("failed to encode evidence: %w", err))
	}

	return bz
}