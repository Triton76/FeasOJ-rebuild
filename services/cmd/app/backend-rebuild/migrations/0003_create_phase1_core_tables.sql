-- 0003_create_phase1_core_tables.sql
-- Phase-1 model baseline tables for backend-rebuild.

CREATE TABLE IF NOT EXISTS users (
  id CHAR(36) PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role ENUM('student', 'teacher', 'admin') NOT NULL,
  avatar VARCHAR(255) NULL,
  synopsis TEXT NULL,
  score INT NOT NULL DEFAULT 0,
  status ENUM('active', 'banned') NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email (email),
  KEY idx_users_role (role),
  KEY idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS classes (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  code VARCHAR(32) NOT NULL,
  description TEXT NULL,
  owner_user_id CHAR(36) NOT NULL,
  status ENUM('active') NOT NULL DEFAULT 'active',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_classes_code (code),
  KEY idx_classes_owner_user_id (owner_user_id),
  KEY idx_classes_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS class_memberships (
  id CHAR(36) PRIMARY KEY,
  class_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  role_in_class ENUM('teacher', 'student') NOT NULL,
  status ENUM('pending', 'active', 'rejected', 'removed') NOT NULL,
  teacher_class_id CHAR(36) GENERATED ALWAYS AS (IF(role_in_class = 'teacher', class_id, NULL)) STORED,
  joined_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_class_memberships_class_user (class_id, user_id),
  UNIQUE KEY uk_class_memberships_teacher_class_id (teacher_class_id),
  KEY idx_class_memberships_user_id (user_id),
  KEY idx_class_memberships_class_role_status (class_id, role_in_class, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS problems (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  content LONGTEXT NOT NULL,
  input TEXT NULL,
  output TEXT NULL,
  difficulty INT NOT NULL,
  time_limit_ms INT NOT NULL,
  memory_limit_mb INT NOT NULL,
  owner_user_id CHAR(36) NOT NULL,
  class_id CHAR(36) NULL,
  visibility ENUM('public', 'class', 'private') NOT NULL,
  status ENUM('draft', 'published', 'archived') NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_problems_owner_user_id (owner_user_id),
  KEY idx_problems_class_id (class_id),
  KEY idx_problems_visibility_status (visibility, status),
  KEY idx_problems_difficulty (difficulty)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS test_cases (
  id CHAR(36) PRIMARY KEY,
  problem_id BIGINT NOT NULL,
  input_data LONGTEXT NOT NULL,
  output_data LONGTEXT NOT NULL,
  is_sample BOOLEAN NOT NULL,
  sort_order INT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_test_cases_problem_id (problem_id),
  KEY idx_test_cases_problem_sample (problem_id, is_sample),
  KEY idx_test_cases_problem_sort (problem_id, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS contest_problems (
  id CHAR(36) PRIMARY KEY,
  contest_id BIGINT NOT NULL,
  problem_id BIGINT NOT NULL,
  display_order INT NOT NULL,
  alias VARCHAR(16) NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL, 
  UNIQUE KEY uk_contest_problems_contest_problem (contest_id, problem_id),
  UNIQUE KEY uk_contest_problems_contest_display_order (contest_id, display_order),
  UNIQUE KEY uk_contest_problems_contest_alias (contest_id, alias)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS discussions (
  id CHAR(36) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content LONGTEXT NOT NULL,
  user_id CHAR(36) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_discussions_user_id (user_id),
  KEY idx_discussions_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS comments (
  id CHAR(36) PRIMARY KEY,
  discussion_id CHAR(36) NOT NULL,
  content LONGTEXT NOT NULL,
  user_id CHAR(36) NOT NULL,
  profanity BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_comments_discussion_id (discussion_id),
  KEY idx_comments_user_id (user_id),
  KEY idx_comments_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS contest_participants (
  id CHAR(36) PRIMARY KEY,
  contest_id BIGINT NOT NULL,
  user_id CHAR(36) NOT NULL,
  status ENUM('registered', 'active', 'quit', 'finished') NOT NULL,
  joined_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_contest_participants_contest_user (contest_id, user_id),
  KEY idx_contest_participants_contest_status (contest_id, status),
  KEY idx_contest_participants_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS submissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id CHAR(36) NOT NULL,
  problem_id BIGINT NOT NULL,
  contest_id BIGINT NULL,
  language VARCHAR(32) NOT NULL,
  source_code LONGTEXT NOT NULL,
  result ENUM(
    'pending',
    'judging',
    'accepted',
    'wrong_answer',
    'compile_error',
    'runtime_error',
    'time_limit_exceeded',
    'memory_limit_exceeded',
    'output_limit_exceeded',
    'presentation_error',
    'partially_accepted',
    'system_error'
  ) NOT NULL,
  score INT NULL,
  submitted_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_submissions_user_id (user_id),
  KEY idx_submissions_problem_id (problem_id),
  KEY idx_submissions_contest_id (contest_id),
  KEY idx_submissions_user_submitted_at (user_id, submitted_at),
  KEY idx_submissions_contest_submitted_at (contest_id, submitted_at),
  KEY idx_submissions_result (result)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO schema_migrations (version, applied_at)
VALUES ('0003', UTC_TIMESTAMP())
ON DUPLICATE KEY UPDATE version = version;