CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(100) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL,
    team_name VARCHAR(100) REFERENCES teams(team_name) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pr_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pr_name VARCHAR(100) NOT NULL,
    author_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(10) NOT NULL CHECK (status IN ('OPEN', 'MERGED')) DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id UUID NOT NULL REFERENCES pull_requests(pr_id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pr_id, reviewer_id)
);
