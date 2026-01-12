package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
)


func NewConversationStore(queries *queries.Queries) *ConversationStore {
	return &ConversationStore{
		queries: queries,
	}
}

type ConversationStore struct {
	queries *queries.Queries
}

func (s *ConversationStore) GetByUserID(ctx context.Context, id uuid.UUID) ([]queries.GetConversationsByUserIDRow, error) {
	conversations, err := s.queries.GetConversationsByUserID(ctx, id)
	return conversations, mapError(err)
}

func (s *ConversationStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.queries.DeleteConversation(ctx, id)
	return mapError(err)
}

func (s *ConversationStore) Create(ctx context.Context, params queries.CreateConversationParams) (queries.Conversation,error) {
	conversation, err := s.queries.CreateConversation(ctx, params)
	return conversation, mapError(err)
}

func (s *ConversationStore) GetByMembers(ctx context.Context, params queries.GetConversationByMembersParams) (queries.Conversation, error) {
	c, err := s.queries.GetConversationByMembers(ctx, params)
	return c, mapError(err)
}
