package service

import (
	"mPR/internal/pkg/storage/repository"
	"mPR/internal/service/pullRequests"
	"mPR/internal/service/teams"
	"mPR/internal/service/users"
)

type Manager struct {
	Teams        *teams.Service
	Users        *users.Service
	PullRequests *pullRequests.Service
}

func New(all *repository.All) *Manager {
	return &Manager{
		Teams:        teams.New(all.Teams, all.Users),
		Users:        users.New(all.Users, all.PullRequests, all.Reviewers),
		PullRequests: pullRequests.New(all.PullRequests, all.Users, all.Reviewers),
	}
}
