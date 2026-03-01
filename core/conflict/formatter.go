// Package conflict 提供冲突检测结果的格式化输出。
package conflict

import (
	"fmt"
	"strings"
)

// Formatter 冲突结果格式化器。
type Formatter struct {
	useColor bool
}

// NewFormatter 创建新的格式化器。
func NewFormatter(useColor bool) *Formatter {
	return &Formatter{useColor: useColor}
}

// Format 格式化冲突检测结果为字符串。
func (f *Formatter) Format(result *Result) string {
	if result == nil || len(result.Conflicts) == 0 {
		return "✓ 未检测到冲突"
	}

	var sb strings.Builder

	// 标题
	sb.WriteString("========================================\n")
	sb.WriteString("冲突检测报告\n")
	sb.WriteString("========================================\n\n")

	// 按类型分组
	grouped := f.groupByType(result.Conflicts)

	// 严重冲突
	if result.HasCritical {
		sb.WriteString("⚠ 发现严重冲突，建议先解决后再安装\n\n")
	} else if result.HasWarning {
		sb.WriteString("⚡ 发现警告级别冲突，可以使用 --force 强制安装\n\n")
	}

	// 输出各类型冲突
	for conflictType, conflicts := range grouped {
		f.formatConflictGroup(&sb, conflictType, conflicts)
	}

	// 汇总
	sb.WriteString("----------------------------------------\n")
	sb.WriteString(fmt.Sprintf("总计: %d 个冲突\n", len(result.Conflicts)))
	sb.WriteString(fmt.Sprintf("  严重: %d\n", f.countBySeverity(result.Conflicts, SeverityCritical)))
	sb.WriteString(fmt.Sprintf("  警告: %d\n", f.countBySeverity(result.Conflicts, SeverityWarning)))
	sb.WriteString(fmt.Sprintf("  信息: %d\n", f.countBySeverity(result.Conflicts, SeverityInfo)))
	sb.WriteString("========================================\n")

	return sb.String()
}

// FormatSimple 简化格式输出。
func (f *Formatter) FormatSimple(result *Result) string {
	if result == nil || len(result.Conflicts) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, c := range result.Conflicts {
		switch c.Severity {
		case SeverityCritical:
			sb.WriteString(fmt.Sprintf("[严重] %s: %s\n", c.Type, c.Description))
		case SeverityWarning:
			sb.WriteString(fmt.Sprintf("[警告] %s: %s\n", c.Type, c.Description))
		case SeverityInfo:
			sb.WriteString(fmt.Sprintf("[信息] %s: %s\n", c.Type, c.Description))
		}
	}
	return sb.String()
}

// groupByType 按类型分组冲突。
func (f *Formatter) groupByType(conflicts []Conflict) map[ConflictType][]Conflict {
	grouped := make(map[ConflictType][]Conflict)
	for _, c := range conflicts {
		grouped[c.Type] = append(grouped[c.Type], c)
	}
	return grouped
}

// formatConflictGroup 格式化冲突组。
func (f *Formatter) formatConflictGroup(sb *strings.Builder, conflictType ConflictType, conflicts []Conflict) {
	if len(conflicts) == 0 {
		return
	}

	// 类型标题
	typeName := f.getTypeDisplayName(conflictType)
	sb.WriteString(fmt.Sprintf("【%s】\n", typeName))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// 输出每个冲突
	for i, c := range conflicts {
		sb.WriteString(fmt.Sprintf("  %d. ", i+1))

		// 严重程度标识
		switch c.Severity {
		case SeverityCritical:
			sb.WriteString("[严重] ")
		case SeverityWarning:
			sb.WriteString("[警告] ")
		case SeverityInfo:
			sb.WriteString("[信息] ")
		}

		sb.WriteString(c.Description + "\n")

		// 当前占用者
		if c.CurrentApp != "" && c.CurrentApp != "unknown" {
			sb.WriteString(fmt.Sprintf("     当前占用: %s\n", c.CurrentApp))
		}

		// 目标
		if c.Target != "" {
			sb.WriteString(fmt.Sprintf("     目标: %s\n", c.Target))
		}

		// 建议
		if c.Suggestion != "" {
			sb.WriteString(fmt.Sprintf("     建议: %s\n", c.Suggestion))
		}

		sb.WriteString("\n")
	}
}

// getTypeDisplayName 获取冲突类型的显示名称。
func (f *Formatter) getTypeDisplayName(t ConflictType) string {
	switch t {
	case ConflictTypeFile:
		return "文件冲突"
	case ConflictTypePort:
		return "端口冲突"
	case ConflictTypeEnvVar:
		return "环境变量冲突"
	case ConflictTypeRegistry:
		return "注册表冲突"
	case ConflictTypeDependency:
		return "依赖冲突"
	default:
		return string(t)
	}
}

// countBySeverity 按严重程度计数。
func (f *Formatter) countBySeverity(conflicts []Conflict, severity Severity) int {
	count := 0
	for _, c := range conflicts {
		if c.Severity == severity {
			count++
		}
	}
	return count
}

// ShouldBlockInstall 判断是否应阻止安装。
func ShouldBlockInstall(result *Result, force bool) bool {
	if result == nil {
		return false
	}

	// 如果有严重冲突且未使用 force，则阻止安装
	if result.HasCritical && !force {
		return true
	}

	return false
}

// HasConflicts 检查是否有冲突。
func HasConflicts(result *Result) bool {
	return result != nil && len(result.Conflicts) > 0
}
