// Package git 提供 Git 操作功能。
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Client struct {
	worktree *git.Worktree
	repo     *git.Repository
}

func Clone(url, dest string, depth int) (*Client, error) {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, fmt.Errorf("创建目标目录: %w", err)
	}

	opts := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	if depth > 0 {
		opts.Depth = depth
	}

	repo, err := git.PlainClone(dest, false, opts)
	if err != nil {
		return nil, fmt.Errorf("克隆仓库: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("获取工作树: %w", err)
	}

	return &Client{
		repo:     repo,
		worktree: worktree,
	}, nil
}

func Open(dir string) (*Client, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("打开仓库: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("获取工作树: %w", err)
	}

	return &Client{
		repo:     repo,
		worktree: worktree,
	}, nil
}

func (c *Client) Pull() error {
	err := c.worktree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("拉取更新: %w", err)
	}
	return nil
}

func (c *Client) Fetch() error {
	err := c.repo.Fetch(&git.FetchOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("获取更新: %w", err)
	}
	return nil
}

func (c *Client) GetLatestTag() (string, error) {
	tags, err := c.repo.Tags()
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

func (c *Client) GetCommitHash() (string, error) {
	ref, err := c.repo.Head()
	if err != nil {
		return "", fmt.Errorf("获取 HEAD: %w", err)
	}
	return ref.Hash().String(), nil
}

func (c *Client) GetCurrentBranch() (string, error) {
	ref, err := c.repo.Head()
	if err != nil {
		return "", fmt.Errorf("获取 HEAD: %w", err)
	}
	return ref.Name().Short(), nil
}

func (c *Client) Checkout(branch string) error {
	err := c.worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branch),
		Create: false,
	})
	if err != nil {
		return fmt.Errorf("切换分支: %w", err)
	}
	return nil
}

func (c *Client) RemoteURL() (string, error) {
	remotes, err := c.repo.Remotes()
	if err != nil {
		return "", err
	}
	if len(remotes) == 0 {
		return "", fmt.Errorf("没有远程仓库")
	}
	cfg, err := c.repo.Config()
	if err != nil {
		return "", err
	}
	for _, r := range remotes {
		if r.Config().Name == "origin" {
			if len(r.Config().URLs) > 0 {
				return r.Config().URLs[0], nil
			}
		}
	}
	_ = cfg
	return "", fmt.Errorf("未找到 origin 远程仓库")
}

func (c *Client) Close() error {
	return nil
}
