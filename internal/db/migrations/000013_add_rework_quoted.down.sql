ALTER TABLE jobs DROP COLUMN IF EXISTS rework_quote_amount;
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check CHECK (status IN (
  'pending_visit','visit_scheduled','visit_quoted','visit_paid',
  'visit_completed','work_quoted','work_approved','work_in_progress',
  'work_delivered','rework_requested','completed','cancelled'
));
