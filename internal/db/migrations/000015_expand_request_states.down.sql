DROP INDEX IF EXISTS uq_requests_active_per_pair;
ALTER TABLE requests DROP CONSTRAINT IF EXISTS requests_status_check;
ALTER TABLE requests ADD CONSTRAINT requests_status_check CHECK (
  status IN ('pending', 'accepted', 'rejected')
);
