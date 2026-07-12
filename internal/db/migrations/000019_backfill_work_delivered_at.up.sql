UPDATE jobs
SET work_delivered_at = NOW()
WHERE status = 'work_delivered' AND work_delivered_at IS NULL;
