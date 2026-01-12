package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myselfBZ/chatrix-v2/internal/queries"
)

func NewStorage(db *pgxpool.Pool) *Storage {
	queries := queries.New(db)
	return &Storage{
		Users: NewUserStore(queries),
		Contacts: NewContactStore(queries),
		Messages: NewMessageStore(queries),
		Conversations: NewConversationStore(queries),
	}
}

type Storage struct {
	Users interface {
		Create(ctx context.Context, arg queries.CreateUserParams) (queries.User, error)

		GetByID(ctx context.Context, id uuid.UUID) (queries.User, error)
		GetByUsername(ctx context.Context, username string) (queries.User, error)
		GetByEmail(ctx context.Context, email string) (queries.User, error)

		List(ctx context.Context) ([]queries.User, error)
		Search(ctx context.Context, username string) ([]queries.User, error)

		UpdateLastSeen(ctx context.Context, id uuid.UUID) error
		// UpdatePassword(ctx context.Context, arg queries.UpdateUserPasswordParams) error
	}

	Contacts interface {
		Add(ctx context.Context, arg queries.AddContactParams) error

		GetByUserID(ctx context.Context, userID uuid.UUID) ([]queries.GetContactsByUserIDRow, error)

		Delete(ctx context.Context, arg queries.DeleteContactParams) error

		// Search(ctx context.Context, arg queries.SearchContactsParams) ([]queries.User, error)
	}


	Messages interface {
		Create(ctx context.Context, arg queries.CreateMessageParams) (queries.Message, error)

		GetByConversationID(ctx context.Context, arg queries.GetMessagesByConversationIDParams) ([]queries.Message, error)

		MarkAsRead(ctx context.Context, arg queries.MarkMessagesAsReadParams) error

		// GetLast(ctx context.Context, userID uuid.UUID) ([]queries.Message, error)

		Delete(ctx context.Context, id uuid.UUID) error
	}


	Conversations interface {
		GetByUserID(ctx context.Context, id uuid.UUID) ([]queries.GetConversationsByUserIDRow, error) 

		Delete(ctx context.Context, id uuid.UUID) error 

		Create(ctx context.Context, params queries.CreateConversationParams) (queries.Conversation,error) 

		GetByMembers(ctx context.Context, params queries.GetConversationByMembersParams) (queries.Conversation, error)
	}
}
