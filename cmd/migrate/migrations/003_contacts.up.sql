CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    contact_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    nickname VARCHAR(50),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT no_self_contact CHECK (user_id <> contact_user_id),
    UNIQUE (user_id, contact_user_id)
);
