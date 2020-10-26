package types

const (
	AttributeValueCategory = ModuleName
)

const (
	EventTypeCreate   = "create"
	EventTypeIssue    = "issue"
	EventTypeBurn     = "burn"
	EventTypeTransfer = "transfer"
	EventTypeLock     = "lock"
	EventTypeUnlock   = "unlock"
	EventTypePayFee   = "payfee"
	EventTypeExercise = "exercise"
	EventTypeApprove  = "approve"
)

const (
	AttributeKeyFrom              = "from"
	AttributeKeyTo                = "to"
	AttributeKeySpender       = "spender"
	AttributeKeyAmount            = "amount"
	AttributeKeyCreator           = "creator"
	AttributeKeySymbol            = "symbol"
	AttributeKeyMaxSupply         = "max"
	AttributeKeySupply            = "supply"
	AttributeKeyAccount           = "id"
	AttributeKeyUnlockHeight      = "unlockHeight"
	AttributeKeyCanIssue          = "canIssue"
	AttributeKeyCanLock           = "canLock"
	AttributeKeyIssueToHeight     = "issueToHeight"
	AttributeKeyIssueCreateHeight = "issueCreateHeight"
	AttributeKeyHeight            = "Height"
	AttributeKeyInit              = "init"
	AttributeKeyDescription       = "desc"
)
