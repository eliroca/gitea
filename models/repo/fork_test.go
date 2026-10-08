// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo_test

import (
	"testing"

	repo_model "gitea.dev/models/repo"
	"gitea.dev/models/unittest"

	"github.com/stretchr/testify/assert"
)

func TestGetUserFork(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	// User13 has repo 11 forked from repo10
	repo, err := repo_model.GetRepositoryByID(t.Context(), 10)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	repo, err = repo_model.GetUserFork(t.Context(), repo.ID, 13)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	repo, err = repo_model.GetRepositoryByID(t.Context(), 9)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	repo, err = repo_model.GetUserFork(t.Context(), repo.ID, 13)
	assert.NoError(t, err)
	assert.Nil(t, repo)
}

func TestReparentFork(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	// Build 9 <- 10 <- 11, then add repo 12 as a sibling fork of 10.
	forkedRepo, err := repo_model.GetRepositoryByID(t.Context(), 11)
	assert.NoError(t, err)
	assert.True(t, forkedRepo.IsFork)
	assert.Equal(t, int64(10), forkedRepo.ForkID)

	parentRepo, err := repo_model.GetRepositoryByID(t.Context(), 10)
	assert.NoError(t, err)
	assert.False(t, parentRepo.IsFork)
	assert.Equal(t, int64(0), parentRepo.ForkID)
	oldParent, err := repo_model.GetRepositoryByID(t.Context(), 9)
	assert.NoError(t, err)
	oldParentForkCount := oldParent.NumForks
	parentRepo.IsFork = true
	parentRepo.ForkID = oldParent.ID
	assert.NoError(t, repo_model.UpdateRepositoryColsNoAutoTime(t.Context(), parentRepo, "is_fork", "fork_id"))
	assert.NoError(t, repo_model.IncrementRepoForkNum(t.Context(), oldParent.ID))

	siblingFork, err := repo_model.GetRepositoryByID(t.Context(), 12)
	assert.NoError(t, err)
	siblingFork.IsFork = true
	siblingFork.ForkID = parentRepo.ID
	assert.NoError(t, repo_model.UpdateRepositoryColsNoAutoTime(t.Context(), siblingFork, "is_fork", "fork_id"))
	assert.NoError(t, repo_model.IncrementRepoForkNum(t.Context(), parentRepo.ID))

	parentRepoBefore, err := repo_model.GetRepositoryByID(t.Context(), parentRepo.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, parentRepoBefore.NumForks)

	// Reparent the middle of 9 <- 10 <- 11, leaving repository 12 as a sibling fork of 10.
	err = repo_model.ReparentFork(t.Context(), 11, 10)
	assert.NoError(t, err)

	// Verify the swap
	forkedRepoAfter, err := repo_model.GetRepositoryByID(t.Context(), 11)
	assert.NoError(t, err)
	assert.False(t, forkedRepoAfter.IsFork)
	assert.Equal(t, int64(0), forkedRepoAfter.ForkID)
	assert.Equal(t, 1, forkedRepoAfter.NumForks)

	parentRepoAfter, err := repo_model.GetRepositoryByID(t.Context(), 10)
	assert.NoError(t, err)
	assert.True(t, parentRepoAfter.IsFork)
	assert.Equal(t, int64(11), parentRepoAfter.ForkID)
	assert.Equal(t, 1, parentRepoAfter.NumForks)

	siblingForkAfter, err := repo_model.GetRepositoryByID(t.Context(), siblingFork.ID)
	assert.NoError(t, err)
	assert.True(t, siblingForkAfter.IsFork)
	assert.Equal(t, parentRepo.ID, siblingForkAfter.ForkID)

	oldParentAfter, err := repo_model.GetRepositoryByID(t.Context(), oldParent.ID)
	assert.NoError(t, err)
	assert.Equal(t, oldParentForkCount, oldParentAfter.NumForks)
}
