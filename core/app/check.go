package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"chopsticks/pkg/errors"
)

type Operation struct {
	Type          string `json:"type"`
	Path          string `json:"path,omitempty"`
	Key           string `json:"key,omitempty"`
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	Target        string `json:"target,omitempty"`
	Link          string `json:"link,omitempty"`
	OriginalValue string `json:"original_value,omitempty"`
}

type OperationsFile struct {
	Version    string      `json:"version"`
	Operations []Operation `json:"operations"`
}

func (m *manager) Check(ctx context.Context, name string, opts CheckOptions) (*CheckResult, error) {
	installed, err := m.storage.GetInstalledApp(ctx, name)
	if err != nil {
		return nil, errors.NewAppNotInstalled(name)
	}

	result := &CheckResult{
		Name:      name,
		Status:    CheckStatusPassed,
		Issues:    []CheckIssue{},
		CheckedAt: time.Now(),
	}

	if opts.CheckPaths || (!opts.CheckEnv && !opts.CheckSymlinks && !opts.CheckFiles) {
		if issues := m.checkPaths(ctx, installed.InstallDir); issues != nil {
			result.Issues = append(result.Issues, issues...)
		}
	}

	if opts.CheckEnv || (!opts.CheckPaths && !opts.CheckSymlinks && !opts.CheckFiles) {
		if issues := m.checkEnvVars(ctx, installed.InstallDir); issues != nil {
			result.Issues = append(result.Issues, issues...)
		}
	}

	if opts.CheckSymlinks || (!opts.CheckPaths && !opts.CheckEnv && !opts.CheckFiles) {
		if issues := m.checkSymlinks(ctx, installed.InstallDir); issues != nil {
			result.Issues = append(result.Issues, issues...)
		}
	}

	if opts.CheckFiles || (!opts.CheckPaths && !opts.CheckEnv && !opts.CheckSymlinks) {
		if issues := m.checkFiles(ctx, installed.InstallDir); issues != nil {
			result.Issues = append(result.Issues, issues...)
		}
	}

	if len(result.Issues) > 0 {
		result.Status = CheckStatusFailed
	}

	return result, nil
}

func (m *manager) checkPaths(ctx context.Context, installDir string) []CheckIssue {
	var issues []CheckIssue

	ops, err := m.loadOperations(installDir)
	if err != nil {
		return issues
	}

	for _, op := range ops {
		if op.Type != "path" {
			continue
		}

		path := op.Path
		if path == "" {
			continue
		}

		if !filepath.IsAbs(path) {
			path = filepath.Join(installDir, op.Path)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			issues = append(issues, CheckIssue{
				Type:    IssueTypePath,
				Message: "PATH 条目不存在",
				Target:  path,
			})
		}
	}

	return issues
}

func (m *manager) checkEnvVars(_ context.Context, installDir string) []CheckIssue {
	var issues []CheckIssue

	ops, err := m.loadOperations(installDir)
	if err != nil {
		return issues
	}

	for _, op := range ops {
		if op.Type != "env" {
			continue
		}

		key := op.Key
		if key == "" {
			continue
		}

		expectedValue := op.Value
		actualValue := os.Getenv(key)

		if actualValue == "" {
			issues = append(issues, CheckIssue{
				Type:    IssueTypeEnv,
				Message: "环境变量未设置",
				Target:  key,
			})
		} else if expectedValue != "" && actualValue != expectedValue {
			issues = append(issues, CheckIssue{
				Type:    IssueTypeEnv,
				Message: "环境变量值不匹配",
				Target:  key,
			})
		}
	}

	return issues
}

func (m *manager) checkSymlinks(ctx context.Context, installDir string) []CheckIssue {
	var issues []CheckIssue

	ops, err := m.loadOperations(installDir)
	if err != nil {
		return issues
	}

	for _, op := range ops {
		if op.Type != "symlink" {
			continue
		}

		linkPath := op.Link
		if linkPath == "" {
			continue
		}

		if !filepath.IsAbs(linkPath) {
			linkPath = filepath.Join(installDir, op.Link)
		}

		link, err := os.Readlink(linkPath)
		if err != nil {
			issues = append(issues, CheckIssue{
				Type:    IssueTypeSymlink,
				Message: "符号链接无法读取",
				Target:  linkPath,
			})
			continue
		}

		targetPath := op.Target
		if targetPath == "" {
			targetPath = link
		}

		if !filepath.IsAbs(targetPath) {
			targetPath = filepath.Join(installDir, targetPath)
		}

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			issues = append(issues, CheckIssue{
				Type:    IssueTypeSymlink,
				Message: "符号链接目标不存在",
				Target:  linkPath,
			})
		}
	}

	return issues
}

func (m *manager) checkFiles(_ context.Context, installDir string) []CheckIssue {
	var issues []CheckIssue

	ops, err := m.loadOperations(installDir)
	if err != nil {
		return issues
	}

	for _, op := range ops {
		if op.Type != "file" {
			continue
		}

		filePath := op.Path
		if filePath == "" {
			continue
		}

		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(installDir, op.Path)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			issues = append(issues, CheckIssue{
				Type:    IssueTypeFile,
				Message: "文件不存在",
				Target:  filePath,
			})
		}
	}

	return issues
}

func (m *manager) loadOperations(installDir string) ([]Operation, error) {
	opsPath := filepath.Join(installDir, "operations.json")

	data, err := os.ReadFile(opsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Operation{}, nil
		}
		return nil, err
	}

	var opsFile OperationsFile
	if err := json.Unmarshal(data, &opsFile); err != nil {
		return nil, err
	}

	return opsFile.Operations, nil
}
