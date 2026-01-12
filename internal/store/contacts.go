package store

import (
	"context"
	"github.com/google/uuid"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
)

type ContactStore struct {
	q *queries.Queries
}

func NewContactStore(q *queries.Queries) *ContactStore {
	return &ContactStore{q: q}
}

func (s *ContactStore) Add(ctx context.Context, arg queries.AddContactParams) error {
	// sqlc generates AddContact from your INSERT statement
	_, err := s.q.AddContact(ctx, arg)
	return mapError(err)
}

func (s *ContactStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]queries.GetContactsByUserIDRow, error) {
	// This should use the JOIN query that returns the User details
	// of the people in the contact list
	users, err := s.q.GetContactsByUserID(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	return users, nil
}

func (s *ContactStore) Delete(ctx context.Context, arg queries.DeleteContactParams) error {
	_, err := s.q.DeleteContact(ctx, arg)
	return mapError(err)
}
