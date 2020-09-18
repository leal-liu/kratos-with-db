package utils

import (
	"encoding/json"
	"fmt"

	accountTypes "github.com/KuChainNetwork/kuchain/x/account/types"
	assetTypes "github.com/KuChainNetwork/kuchain/x/asset/types"
	distributionTypes "github.com/KuChainNetwork/kuchain/x/distribution/types"
	evidenceTypes "github.com/KuChainNetwork/kuchain/x/evidence/types"
	govTypes "github.com/KuChainNetwork/kuchain/x/gov/types"
	slashingType "github.com/KuChainNetwork/kuchain/x/slashing/types"
	stakingTypes "github.com/KuChainNetwork/kuchain/x/staking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
)

// RequireMsgDataJSON must marshal msg data to json format
func RequireMsgDataJSON(cdc *amino.Codec, msg sdk.Msg) (output json.RawMessage) {
	var input []byte
	var ptr interface{}
	switch msg := msg.(type) {
	// account
	case accountTypes.MsgCreateAccount:
		var data accountTypes.MsgCreateAccountData
		input, ptr = msg.Data, &data
	case accountTypes.MsgUpdateAccountAuth:
		var data accountTypes.MsgUpdateAccountAuthData
		input, ptr = msg.Data, &data

	// asset
	case assetTypes.MsgTransfer:
	case assetTypes.MsgCreateCoin:
		var data assetTypes.MsgCreateCoinData
		input, ptr = msg.Data, &data
	case assetTypes.MsgIssueCoin:
		var data assetTypes.MsgIssueCoinData
		input, ptr = msg.Data, &data
	case assetTypes.MsgBurnCoin:
		var data assetTypes.MsgBurnCoinData
		input, ptr = msg.Data, &data
	case assetTypes.MsgLockCoin:
		var data assetTypes.MsgLockCoinData
		input, ptr = msg.Data, &data
	case assetTypes.MsgExerciseCoin:
		var data assetTypes.MsgExerciseCoinData
		input, ptr = msg.Data, &data

	// distribution
	case distributionTypes.MsgSetWithdrawAccountId:
		var data distributionTypes.MsgSetWithdrawAccountIdData
		input, ptr = msg.Data, &data
	case distributionTypes.MsgWithdrawDelegatorReward:
		var data distributionTypes.MsgWithdrawDelegatorRewardData
		input, ptr = msg.Data, &data
	case distributionTypes.MsgWithdrawValidatorCommission:
		var data distributionTypes.MsgWithdrawValidatorCommissionData
		input, ptr = msg.Data, &data

	// evidence
	case evidenceTypes.MsgSubmitEvidenceBase:
	case evidenceTypes.MsgSubmitEvidence:

	// gov
	case govTypes.KuMsgSubmitProposal:
		var data govTypes.MsgSubmitProposalBase
		input, ptr = msg.Data, &data
	case govTypes.KuMsgDeposit:
		var data govTypes.MsgDeposit
		input, ptr = msg.Data, &data
	case govTypes.KuMsgVote:
		var data govTypes.MsgVote
		input, ptr = msg.Data, &data

	// slashing
	case slashingType.KuMsgUnjail:
		var data slashingType.MsgUnjail
		input, ptr = msg.Data, &data

	// staking
	case stakingTypes.KuMsgCreateValidator:
		var data stakingTypes.MsgCreateValidator
		input, ptr = msg.Data, &data
	case stakingTypes.KuMsgDelegate:
		var data stakingTypes.MsgDelegate
		input, ptr = msg.Data, &data
	case stakingTypes.KuMsgEditValidator:
		var data stakingTypes.MsgEditValidator
		input, ptr = msg.Data, &data
	case stakingTypes.KuMsgRedelegate:
		var data stakingTypes.MsgBeginRedelegate
		input, ptr = msg.Data, &data
	case stakingTypes.KuMsgUnbond:
		var data stakingTypes.MsgUndelegate
		input, ptr = msg.Data, &data
	default:
		panic(fmt.Sprint("unknown message:", msg))
	}
	cdc.MustUnmarshalBinaryLengthPrefixed(input, ptr)
	output = cdc.MustMarshalJSON(ptr)
	return
}
