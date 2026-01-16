package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
)

type UserStore struct {
	q *queries.Queries
}

func NewUserStore(q *queries.Queries) *UserStore {
	return &UserStore{q: q}
}

func (s *UserStore) Create(ctx context.Context, arg queries.CreateUserParams) (queries.User, error) {
	user, err := s.q.CreateUser(ctx, arg)
	if err != nil {
		return queries.User{}, mapError(err)
	}
	return user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (queries.User, error) {
	user, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return queries.User{}, mapError(err)
	}
	return user, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (queries.User, error) {
	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return queries.User{}, mapError(err)
	}
	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (queries.User, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return queries.User{}, mapError(err)
	}
	return user, nil
}

func (s *UserStore) List(ctx context.Context) ([]queries.User, error) {
	users, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	return users, nil
}

func (s *UserStore) Search(ctx context.Context, selfUsername,targetUsername string) ([]queries.SearchUsersRow, error) {
	users, err := s.q.SearchUsers(ctx, queries.SearchUsersParams{
		Column1: pgtype.Text{ String: targetUsername, Valid: true },
		Username: selfUsername,
	})

	if err != nil {
		return nil, mapError(err)
	}
	return users, nil
}

func (s *UserStore) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	err := s.q.UpdateUserLastSeen(ctx, id)
	return mapError(err)
}


// UpdatePassword(ctx context.Context, arg queries.UpdateUserPasswordParams) error

