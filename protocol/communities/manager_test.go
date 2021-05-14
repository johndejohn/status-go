package communities

import (
	"bytes"
	"github.com/status-im/status-go/protocol/requests"
	"testing"

	"github.com/golang/protobuf/proto"
	_ "github.com/mutecomm/go-sqlcipher" // require go-sqlcipher that overrides default implementation

	"github.com/stretchr/testify/suite"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/sqlite"
)

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

type ManagerSuite struct {
	suite.Suite
	manager *Manager
}

func (s *ManagerSuite) SetupTest() {
	db, err := sqlite.OpenInMemory()
	s.Require().NoError(err)
	key, err := crypto.GenerateKey()
	s.Require().NoError(err)
	s.Require().NoError(err)
	m, err := NewManager(&key.PublicKey, db, nil, nil)
	s.Require().NoError(err)
	s.Require().NoError(m.Start())
	s.manager = m
}

func (s *ManagerSuite) TestCreateCommunity() {
	description := &protobuf.CommunityDescription{
		Permissions: &protobuf.CommunityPermissions{
			Access: protobuf.CommunityPermissions_NO_MEMBERSHIP,
		},
		Identity: &protobuf.ChatIdentity{
			DisplayName: "status",
			Description: "status community description",
		},
	}

	community, err := s.manager.CreateCommunity(description)
	s.Require().NoError(err)
	s.Require().NotNil(community)

	communities, err := s.manager.All()
	s.Require().NoError(err)
	// Consider status default community
	s.Require().Len(communities, 2)

	actualCommunity := communities[0]
	if bytes.Equal(community.ID(), communities[1].ID()) {
		actualCommunity = communities[1]
	}

	s.Require().Equal(community.ID(), actualCommunity.ID())
	s.Require().Equal(community.PrivateKey(), actualCommunity.PrivateKey())
	s.Require().True(proto.Equal(community.config.CommunityDescription, actualCommunity.config.CommunityDescription))
}

func (s *ManagerSuite) TestUpdateCommunity() {

	//create community
	description := &protobuf.CommunityDescription{
		Permissions: &protobuf.CommunityPermissions{
			Access: protobuf.CommunityPermissions_NO_MEMBERSHIP,
		},
		Identity: &protobuf.ChatIdentity{
			DisplayName: "Test",
			Description: "test community description",
		},
	}

	community, err := s.manager.CreateCommunity(description)
	s.Require().NoError(err)
	s.Require().NotNil(community)

	update := &requests.UpdateCommunity{
		CommunityID:     community.ID(),
		CreateCommunity: requests.CreateCommunity{
			Name:        "Test 2",
			Description: "Test 2 community description",
		},
	}
	updatedCommunity, err := s.manager.UpdateCommunity(update)
	s.Require().NoError(err)
	s.Require().NotNil(updatedCommunity)

	//ensure updated community successfully stored
	communities, err := s.manager.All()
	s.Require().NoError(err)
	// Consider status default community
	s.Require().Len(communities, 2)

	storedCommunity := communities[0]
	if bytes.Equal(community.ID(), communities[1].ID()) {
		storedCommunity = communities[1]
	}

	s.Require().Equal(storedCommunity.ID(), updatedCommunity.ID())
	s.Require().Equal(storedCommunity.PrivateKey(), updatedCommunity.PrivateKey())
	s.Require().Equal(storedCommunity.config.CommunityDescription.Identity.DisplayName, update.CreateCommunity.Name)
	s.Require().Equal(storedCommunity.config.CommunityDescription.Identity.Description, update.CreateCommunity.Description)
}
