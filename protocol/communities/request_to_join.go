package communities

type RequestToJoin struct {
	Clock       uint64 `json:"clock"`
	ENSName     string `json:"ensName,omitempty"`
	ChatID      string `json:"chatId"`
	CommunityID string `json:"communityId"`
}
