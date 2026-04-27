ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check CHECK (status IN (
  'pending_visit','visit_scheduled','visit_quoted','visit_paid',
  'visit_completed','work_quoted','work_approved','work_in_progress',
  'work_delivered','rework_requested','rework_quoted','completed','cancelled'
));
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS rework_quote_amount NUMERIC(12,2);
