package protocol

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/transport"
	"github.com/status-im/status-go/services/mailservers"
)

func (m *Messenger) scheduleSyncChat(chatID string) error {
	return nil
}

func (m *Messenger) calculateMailserverTo() uint32 {
	return uint32(m.getTimesource().GetCurrentTime() / 1000)
}

func (m *Messenger) filtersForChat(chatID string) ([]*transport.Filter, error) {
	filter := m.transport.FilterByChatID(chatID)
	if filter == nil {
		return nil, errors.New("no filter registered for given chat")
	}

	return []*transport.Filter{filter}, nil
}

// Assume is a public chat for now
func (m *Messenger) syncChat(chatID string) (*MessengerResponse, error) {
	filters, err := m.filtersForChat(chatID)
	if err != nil {
		return nil, err
	}
	return m.syncFilters(filters)
}

func (m *Messenger) defaultSyncPeriod() int {
	return int(m.getTimesource().GetCurrentTime()/1000 - 24*60*60)
}

// RequestAllHistoricMessages requests all the historic messages for any topic
func (m *Messenger) RequestAllHistoricMessages() (*MessengerResponse, error) {
	return m.syncFilters(m.transport.Filters())
}

func (m *Messenger) syncFilters(filters []*transport.Filter) (*MessengerResponse, error) {
	response := &MessengerResponse{}
	topicInfo, err := m.mailserversDatabase.Topics()
	if err != nil {
		return nil, err
	}

	topicsData := make(map[string]mailservers.MailserverTopic)
	for _, topic := range topicInfo {
		topicsData[topic.Topic] = topic
	}

	batches := make(map[int]MailserverBatch)

	var syncedChatIDs []string

	to := m.calculateMailserverTo()
	var syncedTopics []mailservers.MailserverTopic
	for _, filter := range filters {
		if !filter.Listen || filter.Ephemeral {
			continue
		}

		var chatID string
		// If the filter has an identity, we use it as a chatID, otherwise is a public chat/community chat filter
		if len(filter.Identity) != 0 {
			chatID = filter.Identity
		} else {
			chatID = filter.ChatID
		}

		syncedChatIDs = append(syncedChatIDs, chatID)

		topicData, ok := topicsData[filter.Topic.String()]
		if !ok {
			topicData = mailservers.MailserverTopic{
				Topic:       filter.Topic.String(),
				LastRequest: m.defaultSyncPeriod(),
			}
		}
		batch, ok := batches[topicData.LastRequest]
		if !ok {
			batch = MailserverBatch{From: uint32(topicData.LastRequest), To: to}
		}

		batch.ChatIDs = append(batch.ChatIDs, chatID)
		batch.Topics = append(batch.Topics, filter.Topic)
		batches[topicData.LastRequest] = batch
		// Set last request to the new `to`
		topicData.LastRequest = int(to)
		syncedTopics = append(syncedTopics, topicData)
	}

	m.logger.Info("syncing topics", zap.Any("batches", batches))
	for _, batch := range batches {
		m.processMailserverBatch(batch)
	}

	err = m.mailserversDatabase.AddTopics(syncedTopics)
	if err != nil {
		return nil, err
	}

	err = m.persistence.SetLastSynced(to, syncedChatIDs)
	if err != nil {
		return nil, err
	}

	for _, id := range syncedChatIDs {
		chat, ok := m.allChats.Load(id)
		if !ok {
			continue
		}
		chat.LastSynced = to
		response.AddChat(chat)
	}

	return response, nil
}

func (m *Messenger) processMailserverBatch(batch MailserverBatch) error {
	m.logger.Info("syncing topic", zap.Any("topic", batch.Topics), zap.Int64("from", int64(batch.From)), zap.Int64("to", int64(batch.To)))
	cursor, err := m.transport.SendMessagesRequestForTopics(context.Background(), m.mailserver, batch.From, batch.To, nil, batch.Topics, true)
	if err != nil {
		return err
	}
	for len(cursor) != 0 {
		m.logger.Info("retrieved cursor", zap.Any("cursor", cursor))

		cursor, err = m.transport.SendMessagesRequest(context.Background(), m.mailserver, batch.From, batch.To, cursor, true)
		if err != nil {
			return err
		}
	}
	m.logger.Info("synced topic", zap.Any("topic", batch.Topics), zap.Int64("from", int64(batch.From)), zap.Int64("to", int64(batch.To)))
	return nil
}

type MailserverBatch struct {
	From    uint32
	To      uint32
	Cursor  string
	Topics  []types.TopicType
	ChatIDs []string
}

func (m *Messenger) RequestHistoricMessagesForFilter(
	ctx context.Context,
	from, to uint32,
	cursor []byte,
	filter *transport.Filter,
	waitForResponse bool,
) ([]byte, error) {
	if m.mailserver == nil {
		return nil, errors.New("no mailserver selected")
	}

	return m.transport.SendMessagesRequestForFilter(ctx, m.mailserver, from, to, cursor, filter, waitForResponse)
}

func (m *Messenger) LoadFilters(filters []*transport.Filter) ([]*transport.Filter, error) {
	return m.transport.LoadFilters(filters)
}

func (m *Messenger) RemoveFilters(filters []*transport.Filter) error {
	return m.transport.RemoveFilters(filters)
}
