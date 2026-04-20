-- 0006_phase4_testcase_submission_indexes.sql
-- Phase-4 additive indexes for testcase ordering and submission metadata queries.

ALTER TABLE test_cases
  ADD UNIQUE KEY uk_test_cases_problem_sort (problem_id, sort_order);

ALTER TABLE submissions
  ADD KEY idx_submissions_contest_problem_user_time (contest_id, problem_id, user_id, submitted_at),
  ADD KEY idx_submissions_problem_result_time (problem_id, result, submitted_at);
