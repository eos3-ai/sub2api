-- 077_add_scheduled_test_round_id.sql
-- Add round_id for scheduled test results so multi-account results from the same
-- scheduler cycle can be aggregated by round.

ALTER TABLE scheduled_test_results
    ADD COLUMN IF NOT EXISTS round_id VARCHAR(64) NOT NULL DEFAULT '';

-- Fast lookup for round-based aggregation (new data).
CREATE INDEX IF NOT EXISTS idx_str_round_started
    ON scheduled_test_results(round_id, started_at DESC)
    WHERE round_id <> '';

