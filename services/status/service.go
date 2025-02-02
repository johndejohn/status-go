package status

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/account/generator"
	"github.com/status-im/status-go/eth-node/types"
)

// Make sure that Service implements node.Service interface.
var _ node.Service = (*Service)(nil)

// WhisperService whisper interface to add key pairs
type WhisperService interface {
	AddKeyPair(key *ecdsa.PrivateKey) (string, error)
}

// AccountManager interface to manage account actions
type AccountManager interface {
	AddressToDecryptedAccount(string, string) (types.Account, *types.Key, error)
	SelectAccount(account.LoginParams) error
	CreateAccount(password string) (mkInfo generator.GeneratedAccountInfo, accountInfo account.Info, mnemonic string, err error)
}

// Service represents our own implementation of status status operations.
type Service struct {
	am AccountManager
	w  WhisperService
}

// New returns a new Service.
func New(w WhisperService) *Service {
	return &Service{w: w}
}

// Protocols returns a new protocols list. In this case, there are none.
func (s *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns a list of new APIs.
func (s *Service) APIs() []rpc.API {

	return []rpc.API{
		{
			Namespace: "movement",
			Version:   "1.0",
			Service:   NewAPI(s),
			Public:    false,
		},
	}
}

// SetAccountManager sets account manager for the API calls.
func (s *Service) SetAccountManager(a AccountManager) {
	s.am = a
}

// Start is run when a service is started.
// It does nothing in this case but is required by `node.Service` interface.
func (s *Service) Start(server *p2p.Server) error {
	return nil
}

// Stop is run when a service is stopped.
// It does nothing in this case but is required by `node.Service` interface.
func (s *Service) Stop() error {
	return nil
}
