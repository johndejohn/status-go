package protocol

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/protocol/common"
	"github.com/status-im/status-go/protocol/communities"
	"github.com/status-im/status-go/protocol/protobuf"
	"github.com/status-im/status-go/protocol/requests"
)

func (m *Messenger) publishOrg(org *communities.Community) error {
	m.logger.Debug("publishing org", zap.String("org-id", org.IDString()), zap.Any("org", org))
	payload, err := org.MarshaledDescription()
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload: payload,
		Sender:  org.PrivateKey(),
		// we don't want to wrap in an encryption layer message
		SkipEncryption: true,
		MessageType:    protobuf.ApplicationMetadataMessage_COMMUNITY_DESCRIPTION,
	}
	_, err = m.processor.SendPublic(context.Background(), org.IDString(), rawMessage)
	return err
}

func (m *Messenger) publishOrgInvitation(org *communities.Community, invitation *protobuf.CommunityInvitation) error {
	m.logger.Debug("publishing org invitation", zap.String("org-id", org.IDString()), zap.Any("org", org))
	pk, err := crypto.DecompressPubkey(invitation.PublicKey)
	if err != nil {
		return err
	}

	payload, err := proto.Marshal(invitation)
	if err != nil {
		return err
	}

	rawMessage := common.RawMessage{
		Payload: payload,
		Sender:  org.PrivateKey(),
		// we don't want to wrap in an encryption layer message
		SkipEncryption: true,
		MessageType:    protobuf.ApplicationMetadataMessage_COMMUNITY_INVITATION,
	}
	_, err = m.processor.SendPrivate(context.Background(), pk, rawMessage)
	return err
}

// handleCommunitiesSubscription handles events from communities
func (m *Messenger) handleCommunitiesSubscription(c chan *communities.Subscription) {

	var lastPublished int64
	// We check every 5 minutes if we need to publish
	ticker := time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case sub, more := <-c:
				if !more {
					return
				}
				if sub.Community != nil {
					err := m.publishOrg(sub.Community)
					if err != nil {
						m.logger.Warn("failed to publish org", zap.Error(err))
					}
				}

				if sub.Invitation != nil {
					err := m.publishOrgInvitation(sub.Community, sub.Invitation)
					if err != nil {
						m.logger.Warn("failed to publish org invitation", zap.Error(err))
					}
				}

				m.logger.Debug("published org")
			case <-ticker.C:
				// If we are not online, we don't even try
				if !m.online() {
					continue
				}

				// If not enough time has passed since last advertisement, we skip this
				if time.Now().Unix()-lastPublished < communityAdvertiseIntervalSecond {
					continue
				}

				orgs, err := m.communitiesManager.Created()
				if err != nil {
					m.logger.Warn("failed to retrieve orgs", zap.Error(err))
				}

				for idx := range orgs {
					org := orgs[idx]
					err := m.publishOrg(org)
					if err != nil {
						m.logger.Warn("failed to publish org", zap.Error(err))
					}
				}

				// set lastPublished
				lastPublished = time.Now().Unix()

			case <-m.quit:
				return

			}
		}
	}()
}

func (m *Messenger) Communities() ([]*communities.Community, error) {
	return m.communitiesManager.All()
}

func (m *Messenger) JoinedCommunities() ([]*communities.Community, error) {
	return m.communitiesManager.Joined()
}

func (m *Messenger) JoinCommunity(communityID string) (*MessengerResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &MessengerResponse{}

	org, err := m.communitiesManager.JoinCommunity(communityID)
	if err != nil {
		return nil, err
	}

	chatIDs := []string{org.IDString()}

	chats := CreateCommunityChats(org, m.getTimesource())

	// Beware don't use `chat` as a reference
	for i, chat := range chats {
		chatIDs = append(chatIDs, chat.ID)
		response.Chats = append(response.Chats, &chats[i])
	}

	// Load transport filters
	filters, err := m.transport.InitPublicFilters(chatIDs)
	if err != nil {
		return nil, err
	}

	response.Filters = filters
	response.Communities = []*communities.Community{org}

	return response, m.saveChats(response.Chats)
}

func (m *Messenger) LeaveCommunity(communityID string) (*MessengerResponse, error) {
	response := &MessengerResponse{}

	org, err := m.communitiesManager.LeaveCommunity(communityID)
	if err != nil {
		return nil, err
	}

	// Make chat inactive
	for chatID := range org.Chats() {
		orgChatID := communityID + chatID
		err := m.DeleteChat(orgChatID)
		if err != nil {
			return nil, err
		}
		response.RemovedChats = append(response.RemovedChats, orgChatID)

		filter, err := m.transport.RemoveFilterByChatID(orgChatID)
		if err != nil {
			return nil, err
		}

		if filter != nil {
			response.RemovedFilters = append(response.RemovedFilters, filter)
		}
	}

	filter, err := m.transport.RemoveFilterByChatID(communityID)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		response.RemovedFilters = append(response.RemovedFilters, filter)
	}

	response.Communities = []*communities.Community{org}
	return response, nil
}

func (m *Messenger) CreateCommunityChat(orgID string, c *protobuf.CommunityChat) (*MessengerResponse, error) {
	org, changes, err := m.communitiesManager.CreateChat(orgID, c)
	if err != nil {
		return nil, err
	}
	var chats []*Chat
	var chatIDs []string
	for chatID, chat := range changes.ChatsAdded {
		c := CreateCommunityChat(org.IDString(), chatID, chat, m.getTimesource())
		chats = append(chats, &c)
		chatIDs = append(chatIDs, c.ID)
	}

	// Load filters
	filters, err := m.transport.InitPublicFilters(chatIDs)
	if err != nil {
		return nil, err
	}

	return &MessengerResponse{
		Communities:      []*communities.Community{org},
		Chats:            chats,
		Filters:          filters,
		CommunityChanges: []*communities.CommunityChanges{changes},
	}, m.saveChats(chats)
}

func (m *Messenger) CreateCommunity(request *requests.CreateCommunity) (*MessengerResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	description, err := request.ToCommunityDescription()
	if err != nil {
		return nil, err
	}

	description.Members = make(map[string]*protobuf.CommunityMember)
	description.Members[common.PubkeyToHex(&m.identity.PublicKey)] = &protobuf.CommunityMember{}

	org, err := m.communitiesManager.CreateCommunity(description)
	if err != nil {
		return nil, err
	}

	return &MessengerResponse{
		Communities: []*communities.Community{org},
	}, nil
}

func (m *Messenger) ExportCommunity(id string) (*ecdsa.PrivateKey, error) {
	return m.communitiesManager.ExportCommunity(id)
}

func (m *Messenger) ImportCommunity(key *ecdsa.PrivateKey) (*MessengerResponse, error) {
	org, err := m.communitiesManager.ImportCommunity(key)
	if err != nil {
		return nil, err
	}

	// Load filters
	filters, err := m.transport.InitPublicFilters([]string{org.IDString()})
	if err != nil {
		return nil, err
	}

	return &MessengerResponse{
		Filters: filters,
	}, nil
}

func (m *Messenger) InviteUserToCommunity(orgID, pkString string) (*MessengerResponse, error) {
	publicKey, err := common.HexToPubkey(pkString)
	if err != nil {
		return nil, err
	}

	org, err := m.communitiesManager.InviteUserToCommunity(orgID, publicKey)
	if err != nil {
		return nil, err
	}

	return &MessengerResponse{
		Communities: []*communities.Community{org},
	}, nil
}

func (m *Messenger) RemoveUserFromCommunity(orgID, pkString string) (*MessengerResponse, error) {
	publicKey, err := common.HexToPubkey(pkString)
	if err != nil {
		return nil, err
	}

	org, err := m.communitiesManager.RemoveUserFromCommunity(orgID, publicKey)
	if err != nil {
		return nil, err
	}

	return &MessengerResponse{
		Communities: []*communities.Community{org},
	}, nil
}
