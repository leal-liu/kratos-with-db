package keeper

import (
	"fmt"

	chainTypes "github.com/KuChainNetwork/kuchain/chain/types"
	"github.com/KuChainNetwork/kuchain/chain/store"
	"github.com/KuChainNetwork/kuchain/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	sk         types.StakingKeeper
	paramspace types.ParamSubspace
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, sk types.StakingKeeper, paramspace types.ParamSubspace) Keeper {
	return Keeper{
		storeKey:   key,
		cdc:        cdc,
		sk:         sk,
		paramspace: paramspace.WithKeyTable(types.ParamKeyTable()),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// AddPubkey sets a address-pubkey relation
func (k Keeper) AddPubkey(ctx sdk.Context, pubkey crypto.PubKey) {
	addr := pubkey.Address()

	pkStr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pubkey)
	if err != nil {
		panic(fmt.Errorf("error while setting address-pubkey relation: %s", addr))
	}

	k.setAddrPubkeyRelation(ctx, addr, pkStr)
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetPubkey(ctx sdk.Context, address crypto.Address) (crypto.PubKey, error) {
	store := store.NewStore(ctx, k.storeKey)

	var pubkey gogotypes.StringValue
	err := k.cdc.UnmarshalBinaryBare(store.Get(types.GetAddrPubkeyRelationKey(address)), &pubkey)
	if err != nil {
		return nil, fmt.Errorf("address %s not found", sdk.ConsAddress(address))
	}

	pkStr, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pubkey.Value)
	if err != nil {
		return pkStr, err
	}

	return pkStr, nil
}

// Slash attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes.
func (k Keeper) Slash(ctx sdk.Context, consAcc sdk.ConsAddress, fraction sdk.Dec, power, distributionHeight int64) {
	ctx.EventManager().EmitEvent(
		chainTypes.NewEvent(ctx,
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyAddress, consAcc.String()),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
			sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueDoubleSign),
		),
	)

	k.sk.Slash(ctx, consAcc, distributionHeight, power, fraction)
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k Keeper) Jail(ctx sdk.Context, consAcc sdk.ConsAddress) {
	ctx.EventManager().EmitEvent(
		chainTypes.NewEvent(ctx,
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyJailed, consAcc.String()),
		),
	)

	k.sk.Jail(ctx, consAcc)
}

func (k Keeper) setAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address, pubkey string) {
	store := store.NewStore(ctx, k.storeKey)

	bz := k.cdc.MustMarshalBinaryBare(&gogotypes.StringValue{Value: pubkey})
	store.Set(types.GetAddrPubkeyRelationKey(addr), bz)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address) {
	store := store.NewStore(ctx, k.storeKey)
	store.Delete(types.GetAddrPubkeyRelationKey(addr))
}
