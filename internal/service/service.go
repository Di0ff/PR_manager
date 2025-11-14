package service

import (
	"mPR/internal/service/pull_requests"
	"mPR/internal/service/teams"
	"mPR/internal/service/users"
	"mPR/internal/storage/repository"
)

type Manager struct {
	Teams        *teams.Service
	Users        *users.Service
	PullRequests *pull_requests.Service
}

func New(all *repository.All, maxReviewers int) *Manager {
	return &Manager{
		Teams:        teams.New(all.Teams, all.Users),
		Users:        users.New(all.Users, all.PullRequests, all.Reviewers),
		PullRequests: pull_requests.New(all.PullRequests, all.Users, all.Reviewers, maxReviewers),
	}
}
