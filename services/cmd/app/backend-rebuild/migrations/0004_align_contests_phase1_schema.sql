-- 0004_align_contests_phase1_schema.sql
-- Align contests table with phase-1 model draft.

ALTER TABLE contests
  ADD COLUMN subtitle VARCHAR(255) NULL AFTER title,
  ADD COLUMN description LONGTEXT NULL AFTER subtitle,
  ADD COLUMN announcement LONGTEXT NULL AFTER description,
  ADD COLUMN owner_user_id CHAR(36) NOT NULL DEFAULT '' AFTER announcement,
  ADD COLUMN class_id CHAR(36) NULL AFTER owner_user_id,
  ADD COLUMN visibility ENUM('public', 'class', 'private') NOT NULL DEFAULT 'private' AFTER class_id,
  ADD COLUMN rule_type ENUM('acm', 'oi', 'assignment') NOT NULL DEFAULT 'acm' AFTER visibility,
  ADD COLUMN is_encrypted BOOLEAN NOT NULL DEFAULT FALSE AFTER status,
  ADD COLUMN password_hash VARCHAR(255) NULL AFTER is_encrypted,
  ADD COLUMN auto_score BOOLEAN NOT NULL DEFAULT TRUE AFTER password_hash;

ALTER TABLE contests
  ADD INDEX idx_contests_owner_user_id (owner_user_id),
  ADD INDEX idx_contests_class_id (class_id),
  ADD INDEX idx_contests_visibility_status (visibility, status),
  ADD INDEX idx_contests_rule_type (rule_type),
  ADD INDEX idx_contests_start_end (start_at, end_at);

INSERT INTO schema_migrations (version, applied_at)
VALUES ('0004', UTC_TIMESTAMP())
ON DUPLICATE KEY UPDATE version = version;