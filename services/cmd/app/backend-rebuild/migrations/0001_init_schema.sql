-- 0001_init_schema.sql
-- Example baseline migration for a fresh database.

CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(32) PRIMARY KEY,
  applied_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS contests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  status ENUM('draft','scheduled','running','ended') NOT NULL DEFAULT 'draft',
  start_at DATETIME NULL,
  end_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_contests_status (status)
);

INSERT INTO schema_migrations (version, applied_at)
VALUES ('0001', UTC_TIMESTAMP())
ON DUPLICATE KEY UPDATE version = version;
