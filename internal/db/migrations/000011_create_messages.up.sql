CREATE TABLE messages (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id             UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  sender_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  sender_name            TEXT NOT NULL,
  content                TEXT NOT NULL,
  is_unread_for_client   BOOLEAN NOT NULL DEFAULT TRUE,
  is_unread_for_provider BOOLEAN NOT NULL DEFAULT TRUE,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_request_id ON messages(request_id);
