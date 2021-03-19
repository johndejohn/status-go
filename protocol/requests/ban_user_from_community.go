package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/types"
)

var ErrBanUserFromCommunityInvalidCommunityID = errors.New("ban-user-from-community: invalid community id")
var ErrBanUserFromCommunityInvalidUser = errors.New("ban-user-from-community: invalid user id")

type BanUserFromCommunity struct {
	CommunityID types.HexBytes `json:"communityId"`
	User        types.HexBytes `json:"user"`
}

func (j *BanUserFromCommunity) Validate() error {
	if len(j.CommunityID) == 0 {
		return ErrBanUserFromCommunityInvalidCommunityID
	}

	if len(j.User) == 0 {
		return ErrBanUserFromCommunityInvalidUser
	}

	return nil
}
