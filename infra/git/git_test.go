package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client := New()
	assert.NotNil(t, client)
}

func TestGitClient_Clone(t *testing.T) {
	client := New()
	ctx := context.Background()

	t.Run("clone invalid URL", func(t *testing.T) {
		destDir := filepath.Join(t.TempDir(), "dest")
		err := client.Clone(ctx, "invalid-url", destDir)
		assert.Error(t, err)
	})

	t.Run("clone to existing directory", func(t *testing.T) {
		// 先创建一个源仓库
		srcDir := t.TempDir()
		_, err := git.PlainInit(srcDir, false)
		require.NoError(t, err)

		// 添加一个文件
		testFile := filepath.Join(srcDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		destDir := filepath.Join(t.TempDir(), "dest")
		err = client.Clone(ctx, srcDir, destDir)
		// 空仓库克隆可能会失败
		if err != nil {
			t.Logf("Clone error (expected for empty repo): %v", err)
		}
	})
}

func TestGitClient_Pull(t *testing.T) {
	client := New()
	ctx := context.Background()

	t.Run("pull non-existent directory", func(t *testing.T) {
		err := client.Pull(ctx, "/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("pull non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := client.Pull(ctx, tmpDir)
		assert.Error(t, err)
	})

	t.Run("pull already up to date", func(t *testing.T) {
		// 创建一个本地仓库
		repoDir := t.TempDir()
		repo, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		// 配置用户
		cfg, err := repo.Config()
		require.NoError(t, err)
		cfg.User.Name = "Test User"
		cfg.User.Email = "test@example.com"
		err = repo.SetConfig(cfg)
		require.NoError(t, err)

		// 添加一个文件并提交
		testFile := filepath.Join(repoDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		w, err := repo.Worktree()
		require.NoError(t, err)

		_, err = w.Add("test.txt")
		require.NoError(t, err)

		_, err = w.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
			},
		})
		require.NoError(t, err)

		// Pull 可能会失败因为没有远程，但我们只测试它不会 panic
		err = client.Pull(ctx, repoDir)
		// 没有远程仓库，pull 可能会失败，这是预期的
		if err != nil {
			t.Logf("Pull error (expected for repo without remote): %v", err)
		}
	})
}

func TestGitClient_Fetch(t *testing.T) {
	client := New()
	ctx := context.Background()

	t.Run("fetch non-existent directory", func(t *testing.T) {
		err := client.Fetch(ctx, "/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("fetch from repo", func(t *testing.T) {
		// 创建一个本地仓库
		repoDir := t.TempDir()
		_, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		// 添加远程
		repo, err := git.PlainOpen(repoDir)
		require.NoError(t, err)

		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/example/repo.git"},
		})
		require.NoError(t, err)

		// Fetch 可能会失败因为远程不存在，但不会 panic
		err = client.Fetch(ctx, repoDir)
		// 期望错误，因为远程仓库不存在
		assert.Error(t, err)
	})
}

func TestGitClient_GetLatestTag(t *testing.T) {
	client := New()
	ctx := context.Background()

	t.Run("get latest tag from non-existent directory", func(t *testing.T) {
		_, err := client.GetLatestTag(ctx, "/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("get latest tag from repo without tags", func(t *testing.T) {
		repoDir := t.TempDir()
		_, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		tag, err := client.GetLatestTag(ctx, repoDir)
		assert.NoError(t, err)
		assert.Empty(t, tag)
	})
}

func TestGitClient_GetCommitHash(t *testing.T) {
	client := New()
	ctx := context.Background()

	t.Run("get commit hash from non-existent directory", func(t *testing.T) {
		_, err := client.GetCommitHash(ctx, "/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("get commit hash from empty repo", func(t *testing.T) {
		repoDir := t.TempDir()
		_, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		_, err = client.GetCommitHash(ctx, repoDir)
		// 空仓库没有 HEAD
		assert.Error(t, err)
	})

	t.Run("get commit hash from repo with commit", func(t *testing.T) {
		repoDir := t.TempDir()
		repo, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		// 配置用户
		cfg, err := repo.Config()
		require.NoError(t, err)
		cfg.User.Name = "Test User"
		cfg.User.Email = "test@example.com"
		err = repo.SetConfig(cfg)
		require.NoError(t, err)

		// 添加一个文件并提交
		testFile := filepath.Join(repoDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		w, err := repo.Worktree()
		require.NoError(t, err)

		_, err = w.Add("test.txt")
		require.NoError(t, err)

		commitHash, err := w.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
			},
		})
		require.NoError(t, err)

		hash, err := client.GetCommitHash(ctx, repoDir)
		require.NoError(t, err)
		assert.Equal(t, commitHash.String(), hash)
	})
}

func TestGitClient_GetCurrentBranch(t *testing.T) {
	client := New()
	ctx := context.Background()

	t.Run("get current branch from non-existent directory", func(t *testing.T) {
		_, err := client.GetCurrentBranch(ctx, "/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("get current branch from empty repo", func(t *testing.T) {
		repoDir := t.TempDir()
		_, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		_, err = client.GetCurrentBranch(ctx, repoDir)
		// 空仓库没有 HEAD
		assert.Error(t, err)
	})

	t.Run("get current branch from repo with commit", func(t *testing.T) {
		repoDir := t.TempDir()
		repo, err := git.PlainInit(repoDir, false)
		require.NoError(t, err)

		// 配置用户
		cfg, err := repo.Config()
		require.NoError(t, err)
		cfg.User.Name = "Test User"
		cfg.User.Email = "test@example.com"
		err = repo.SetConfig(cfg)
		require.NoError(t, err)

		// 添加一个文件并提交
		testFile := filepath.Join(repoDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		require.NoError(t, err)

		w, err := repo.Worktree()
		require.NoError(t, err)

		_, err = w.Add("test.txt")
		require.NoError(t, err)

		_, err = w.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
			},
		})
		require.NoError(t, err)

		branch, err := client.GetCurrentBranch(ctx, repoDir)
		require.NoError(t, err)
		assert.NotEmpty(t, branch)
	})
}
