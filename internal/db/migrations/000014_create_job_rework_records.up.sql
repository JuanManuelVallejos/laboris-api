CREATE TABLE job_rework_records (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_id       UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  cycle_number INT  NOT NULL,
  notes        TEXT,
  quote_amount NUMERIC(12,2),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (job_id, cycle_number)
);
CREATE INDEX idx_job_rework_records_job_id ON job_rework_records(job_id);
