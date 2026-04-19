-- 0002_add_contest_status_scan_index.sql
-- Example incremental migration for scheduler scan performance.

ALTER TABLE contests
  ADD INDEX idx_contests_status_start_end (status, start_at, end_at);

INSERT INTO schema_migrations (version, applied_at)
VALUES ('0002', UTC_TIMESTAMP())
ON DUPLICATE KEY UPDATE version = version;
