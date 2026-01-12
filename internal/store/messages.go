package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
)

type MessageStore struct {
	q *queries.Queries
}

func NewMessageStore(q *queries.Queries) *MessageStore {
	return &MessageStore{q: q}
}

func (s *MessageStore) Create(ctx context.Context, arg queries.CreateMessageParams) (queries.Message, error) {
	msg, err := s.q.CreateMessage(ctx, arg)
	if err != nil {
		return queries.Message{}, mapError(err)
	}
	return msg, nil
}

func (s *MessageStore) GetByConversationID(ctx context.Context, arg queries.GetMessagesByConversationIDParams) ([]queries.Message, error) {
	msgs, err := s.q.GetMessagesByConversationID(ctx, arg)
	if err != nil {
		return nil, mapError(err)
	}
	return msgs, nil
}

func (s *MessageStore) MarkAsRead(ctx context.Context, arg queries.MarkMessagesAsReadParams) error {
	err := s.q.MarkMessagesAsRead(ctx, arg)
	return mapError(err)
}

func (s *MessageStore) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.q.DeleteMessage(ctx, id)
	return mapError(err)
}
