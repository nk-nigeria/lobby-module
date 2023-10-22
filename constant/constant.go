package constant

const (
	UUID_USER_SYSTEM = "00000000-0000-0000-0000-000000000000"
)

const (
	RESET_SCHEDULER_LEADER_BOARD = "0 0 * * 1" // At 00:00 on Monday.
)

const NastEndpoint = "nats://nats:Admin123@103.226.250.195:4222"

type UserGroupType string

const UserGroupType_All UserGroupType = "all"
const UserGroupType_Level UserGroupType = "level"
const UserGroupType_VipLevel UserGroupType = "vip_level"
const UserGroupType_WalletChips UserGroupType = "wallet_chips_amount"
const UserGroupType_WalletChipsInbank UserGroupType = "wallet_chips_in_bank_amount"
const UserGroupType_TotalCashOut UserGroupType = "total_cashout"
const UserGroupType_TotalCashOutInDay UserGroupType = "total_cashout_in_day"
const UserGroupType_TotalCashIn UserGroupType = "total_cashin"
const UserGroupType_TotalCashIn1Day UserGroupType = "total_cashin_1_day"
const UserGroupType_TotalCashIn3Day UserGroupType = "total_cashin_3_day"
const UserGroupType_TotalCashIn5Day UserGroupType = "total_cashin_5_day"
const UserGroupType_TotalCashIn7Day UserGroupType = "total_cashin_7_day"
