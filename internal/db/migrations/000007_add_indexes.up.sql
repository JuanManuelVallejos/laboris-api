CREATE INDEX IF NOT EXISTS idx_requests_professional_id ON requests(professional_id);
CREATE INDEX IF NOT EXISTS idx_requests_client_id      ON requests(client_id);
CREATE INDEX IF NOT EXISTS idx_requests_status         ON requests(status);
