-- 0008_align_legacy_collations.sql
-- Align legacy tables created before phase-1 with the modern utf8mb4_0900_ai_ci collation.

ALTER TABLE schema_migrations
  CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

ALTER TABLE contests
  CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

INSERT INTO schema_migrations (version, applied_at)
VALUES ('0008', UTC_TIMESTAMP())
ON DUPLICATE KEY UPDATE version = version;
