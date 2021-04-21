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

// Assume is a public chat for now
func (m *Messenger) syncChat(chatID string) error {
	filter := m.transport.FilterByChatID(chatID)
	if filter == nil {
		return errors.New("no filter registered for given chat")
	}
	to := m.calculateMailserverTo()
	batch := MailserverBatch{From: uint32(m.defaultSyncPeriod()), To: to}
	err := m.processMailserverBatch(batch)
	if err != nil {
		return err
	}

	syncedTopics := []mailservers.MailserverTopic{
		{
			ChatIDs:     []string{chatID},
			Topic:       filter.Topic.String(),
			LastRequest: int(to),
		},
	}
	return m.mailserversDatabase.AddTopics(syncedTopics)
}

func (m *Messenger) defaultSyncPeriod() int {
	return int(m.getTimesource().GetCurrentTime()/1000 - 24*60*60)
}

// RequestAllHistoricMessages requests all the historic messages for any topic
func (m *Messenger) RequestAllHistoricMessages() error {
	allFilters := m.transport.Filters()
	topicsToQuery := make(map[string]types.TopicType)

	for _, f := range allFilters {
		if f.Listen && !f.Ephemeral {
			topicsToQuery[f.Topic.String()] = f.Topic
		}
	}

	topicInfo, err := m.mailserversDatabase.Topics()
	if err != nil {
		return err
	}

	topicsData := make(map[string]mailservers.MailserverTopic)
	for _, topic := range topicInfo {
		topicsData[topic.Topic] = topic
	}

	batches := make(map[int]MailserverBatch)

	to := m.calculateMailserverTo()
	var syncedTopics []mailservers.MailserverTopic
	for _, topic := range topicsToQuery {
		topicData, ok := topicsData[topic.String()]
		if !ok {
			topicData = mailservers.MailserverTopic{
				Topic:       topic.String(),
				LastRequest: m.defaultSyncPeriod(),
			}
		}
		batch, ok := batches[topicData.LastRequest]
		if !ok {
			batch = MailserverBatch{From: uint32(topicData.LastRequest), To: to}
		}

		batch.Topics = append(batch.Topics, topic)
		batches[topicData.LastRequest] = batch
		// Set last request to the new `to`
		topicData.LastRequest = int(to)
		syncedTopics = append(syncedTopics, topicData)
	}

	m.logger.Info("syncing topics", zap.Any("batches", batches))
	for _, batch := range batches {
		m.processMailserverBatch(batch)
	}

	return m.mailserversDatabase.AddTopics(syncedTopics)
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
	From   uint32
	To     uint32
	Cursor string
	Topics []types.TopicType
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
