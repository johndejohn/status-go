package protobuf

import (
	"crypto/sha256"

	"github.com/vacp2p/mvds/state"
)

// ID creates the MessageID for a Message
func (m Message) ID() state.MessageID {

	return sha256.Sum256(m.Body)
}
