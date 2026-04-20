-- 0005_phase3_class_role_and_status.sql
-- Phase-3: align class domain status/role enums with teacher-class closure design.

ALTER TABLE classes
  MODIFY COLUMN status ENUM('active', 'archived') NOT NULL DEFAULT 'active';

ALTER TABLE class_memberships
  MODIFY COLUMN role_in_class ENUM('teacher', 'assistant', 'student') NOT NULL,
  MODIFY COLUMN status ENUM('pending', 'active', 'rejected', 'revoked') NOT NULL;
