
package requests

import (
    "errors"
    "github.com/status-im/status-go/eth-node/types"

    "github.com/status-im/status-go/protocol/protobuf"
)

var (
    ErrUpdateCommunityInvalidID     = errors.New("update-community: invalid id")
    ErrUpdateCommunityInvalidName        = errors.New("update-community: invalid name")
    ErrUpdateCommunityInvalidColor       = errors.New("update-community: invalid color")
    ErrUpdateCommunityInvalidDescription = errors.New("update-community: invalid description")
    ErrUpdateCommunityInvalidMembership  = errors.New("update-community: invalid membership")
)

type UpdateCommunity struct {
    CommunityID types.HexBytes
    CreateCommunity
}

func (u *UpdateCommunity) Validate() error {

    if len(u.CommunityID) == 0 {
        return ErrUpdateCommunityInvalidID
    }

    if u.Name == "" {
        return ErrUpdateCommunityInvalidName
    }

    if u.Description == "" {
        return ErrUpdateCommunityInvalidDescription
    }

    if u.Membership == protobuf.CommunityPermissions_UNKNOWN_ACCESS {
        return ErrUpdateCommunityInvalidMembership
    }

    if u.Color == "" {
        return ErrUpdateCommunityInvalidColor
    }

    return nil
}
