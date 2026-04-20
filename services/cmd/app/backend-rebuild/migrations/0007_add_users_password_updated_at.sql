ALTER TABLE users
  ADD COLUMN password_updated_at DATETIME NULL AFTER password_hash;

UPDATE users
SET password_updated_at = COALESCE(password_updated_at, updated_at, created_at);

ALTER TABLE users
  MODIFY COLUMN password_updated_at DATETIME NOT NULL;
