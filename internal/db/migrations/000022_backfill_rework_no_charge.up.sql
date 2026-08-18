UPDATE job_rework_records r
SET no_charge = true
WHERE r.quote_amount IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM jobs j
    WHERE j.id = r.job_id
      AND j.status = 'rework_requested'
      AND j.rework_count = r.cycle_number
  );
