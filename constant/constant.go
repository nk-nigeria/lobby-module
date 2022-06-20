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
