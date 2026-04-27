CREATE TABLE payments (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_id       UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  type         TEXT NOT NULL CHECK (type IN ('visit','work')),
  amount       NUMERIC(12,2) NOT NULL,
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','released','refunded')),
  provider     TEXT NOT NULL DEFAULT 'mock',
  provider_ref TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (job_id, type)
);

CREATE INDEX idx_payments_job_id ON payments(job_id);
