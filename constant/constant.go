package constant

const (
	UUID_USER_SYSTEM = "00000000-0000-0000-0000-000000000000"
)

const (
	RESET_SCHEDULER_LEADER_BOARD = "0 0 * * 1" // At 00:00 on Monday.
)

const NastEndpoint = "nats://nats:Admin123@103.226.250.195:4222"

const (
	MinLvAllowUseBank       = 2
	MaxChipAllowAdd   int64 = 9000000000 // 9B
)

type UserGroupType string

const UserGroupType_All UserGroupType = "all"
const UserGroupType_Level UserGroupType = "Level"
const UserGroupType_VipLevel UserGroupType = "Vip"
const UserGroupType_WalletChips UserGroupType = "AG"
const UserGroupType_WalletChipsInbank UserGroupType = "wallet_chips_in_bank_amount"
const UserGroupType_TotalCashOut UserGroupType = "CO"
const UserGroupType_TotalCashOutInDay UserGroupType = "CO0"
const UserGroupType_TotalCashIn UserGroupType = "LQ"
const UserGroupType_TotalCashIn1Day UserGroupType = "BLQ1"
const UserGroupType_TotalCashIn3Day UserGroupType = "BLQ3"
const UserGroupType_TotalCashIn5Day UserGroupType = "BLQ5"
const UserGroupType_TotalCashIn7Day UserGroupType = "BLQ7"
const UserGroupType_AvgCashIn7Day UserGroupType = "Avgtrans7"
const UserGroupType_CreateTime UserGroupType = "CreateTime"
