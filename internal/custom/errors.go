package custom

import "errors"

var (
	ErrTeamExists  = errors.New("TEAM_EXISTS")
	ErrPRExists    = errors.New("PR_EXISTS")
	ErrNotFound    = errors.New("NOT_FOUND")
	ErrPRMerged    = errors.New("PR_MERGED")
	ErrNotAssigned = errors.New("NOT_ASSIGNED")
	ErrNoCandidate = errors.New("NO_CANDIDATE")
)
