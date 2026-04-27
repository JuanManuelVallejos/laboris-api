CREATE TABLE jobs (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id         UUID NOT NULL UNIQUE REFERENCES requests(id) ON DELETE CASCADE,
  client_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  professional_id    UUID NOT NULL REFERENCES professionals(id) ON DELETE CASCADE,
  status             TEXT NOT NULL DEFAULT 'pending_visit' CHECK (status IN (
    'pending_visit','visit_scheduled','visit_quoted','visit_paid',
    'visit_completed','work_quoted','work_approved','work_in_progress',
    'work_delivered','rework_requested','completed','cancelled'
  )),
  visit_scheduled_at TIMESTAMPTZ,
  visit_quote_amount NUMERIC(12,2),
  work_quote_amount  NUMERIC(12,2),
  work_description   TEXT,
  rework_count       INT NOT NULL DEFAULT 0,
  rework_notes       TEXT,
  cancel_reason      TEXT,
  completed_at       TIMESTAMPTZ,
  cancelled_at       TIMESTAMPTZ,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_client_id       ON jobs(client_id);
CREATE INDEX idx_jobs_professional_id ON jobs(professional_id);
CREATE INDEX idx_jobs_status          ON jobs(status);
