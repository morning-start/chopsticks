package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Git interface {
	Clone(ctx context.Context, url, dest string) error
	Pull(ctx context.Context, dir string) error
	Fetch(ctx context.Context, dir string) error
	GetLatestTag(ctx context.Context, dir string) (string, error)
	GetCommitHash(ctx context.Context, dir string) (string, error)
	GetCurrentBranch(ctx context.Context, dir string) (string, error)
}

type gitClient struct{}

func New() Git {
	return &gitClient{}
}

func (g *gitClient) Clone(ctx context.Context, url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	opts := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	_, err := git.PlainClone(dest, false, opts)
	if err != nil {
		return fmt.Errorf("克隆仓库: %w", err)
	}

	return nil
}

func (g *gitClient) Pull(ctx context.Context, dir string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("打开仓库: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作树: %w", err)
	}

	err = worktree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("拉取更新: %w", err)
	}
	return nil
}

func (g *gitClient) Fetch(ctx context.Context, dir string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("打开仓库: %w", err)
	}

	err = repo.Fetch(&git.FetchOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("获取更新: %w", err)
	}
	return nil
}

func (g *gitClient) GetLatestTag(ctx context.Context, dir string) (string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", fmt.Errorf("打开仓库: %w", err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("获取标签: %w", err)
	}

	var latestTag string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().String()
		if strings.HasPrefix(name, "refs/tags/") {
			tag := strings.TrimPrefix(name, "refs/tags/")
			if latestTag == "" || tag > latestTag {
				latestTag = tag
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return latestTag, nil
}

func (g *gitClient) GetCommitHash(ctx context.Context, dir string) (string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", fmt.Errorf("打开仓库: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("获取 HEAD: %w", err)
	}
	return ref.Hash().String(), nil
}

func (g *gitClient) GetCurrentBranch(ctx context.Context, dir string) (string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", fmt.Errorf("打开仓库: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("获取 HEAD: %w", err)
	}
	return ref.Name().Short(), nil
}
