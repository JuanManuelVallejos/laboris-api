CREATE TABLE attachments (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type         TEXT NOT NULL,
  owner_id     UUID NOT NULL,
  path         TEXT NOT NULL,
  filename     TEXT NOT NULL,
  extension    TEXT NOT NULL,
  uploaded_by  UUID NOT NULL REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_attachments_type_owner ON attachments(type, owner_id);
