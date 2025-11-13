CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_reviewer_user ON pr_reviewers(reviewer_id);