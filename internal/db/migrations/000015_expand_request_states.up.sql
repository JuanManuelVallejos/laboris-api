ALTER TABLE requests DROP CONSTRAINT IF EXISTS requests_status_check;
ALTER TABLE requests ADD CONSTRAINT requests_status_check CHECK (
  status IN ('pending', 'viewed', 'accepted', 'rejected', 'expired')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_requests_active_per_pair
  ON requests (client_id, professional_id)
  WHERE status IN ('pending', 'viewed');
