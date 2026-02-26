// Package semver 提供语义化版本处理功能。
package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version 表示语义化版本。
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string // 预发布版本
	Build string // 构建元数据
}

// 版本解析正则表达式
var versionRegex = regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z-.]+))?(?:\+([0-9A-Za-z-.]+))?$`)

// Parse 解析版本字符串。
func Parse(version string) (*Version, error) {
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	matches := versionRegex.FindStringSubmatch(version)
	if matches == nil {
		return nil, fmt.Errorf("无效的版本格式: %s", version)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	if matches[2] == "" {
		minor = 0
	}
	patch, _ := strconv.Atoi(matches[3])
	if matches[3] == "" {
		patch = 0
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   matches[4],
		Build: matches[5],
	}, nil
}

// String 返回版本字符串表示。
func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare 比较两个版本。
// 返回 -1 表示 v < other，0 表示相等，1 表示 v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// 处理预发布版本
	if v.Pre == "" && other.Pre != "" {
		return 1 // 正式版本 > 预发布版本
	}
	if v.Pre != "" && other.Pre == "" {
		return -1
	}
	if v.Pre != "" && other.Pre != "" {
		return comparePre(v.Pre, other.Pre)
	}

	return 0
}

// comparePre 比较预发布版本。
func comparePre(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aNum, aErr := strconv.Atoi(aParts[i])
		bNum, bErr := strconv.Atoi(bParts[i])

		// 都是数字，按数字比较
		if aErr == nil && bErr == nil {
			if aNum != bNum {
				if aNum < bNum {
					return -1
				}
				return 1
			}
			continue
		}

		// 数字 < 非数字
		if aErr == nil && bErr != nil {
			return -1
		}
		if aErr != nil && bErr == nil {
			return 1
		}

		// 都是非数字，按字符串比较
		cmp := strings.Compare(aParts[i], bParts[i])
		if cmp != 0 {
			return cmp
		}
	}

	// 字段少的版本更小
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}
	return 0
}

// GT 检查 v > other。
func (v *Version) GT(other *Version) bool {
	return v.Compare(other) > 0
}

// LT 检查 v < other。
func (v *Version) LT(other *Version) bool {
	return v.Compare(other) < 0
}

// EQ 检查 v == other。
func (v *Version) EQ(other *Version) bool {
	return v.Compare(other) == 0
}

// GTE 检查 v >= other。
func (v *Version) GTE(other *Version) bool {
	return v.Compare(other) >= 0
}

// LTE 检查 v <= other。
func (v *Version) LTE(other *Version) bool {
	return v.Compare(other) <= 0
}

// CompareStrings 比较两个版本字符串。
func CompareStrings(v1, v2 string) (int, error) {
	ver1, err := Parse(v1)
	if err != nil {
		return 0, fmt.Errorf("解析版本1: %w", err)
	}
	ver2, err := Parse(v2)
	if err != nil {
		return 0, fmt.Errorf("解析版本2: %w", err)
	}
	return ver1.Compare(ver2), nil
}

// IsValid 检查版本字符串是否有效。
func IsValid(version string) bool {
	_, err := Parse(version)
	return err == nil
}

// Satisfies 检查版本是否满足约束。
func Satisfies(version, constraint string) (bool, error) {
	v, err := Parse(version)
	if err != nil {
		return false, err
	}

	// 处理简单约束
	constraint = strings.TrimSpace(constraint)

	// >= x.y.z
	if strings.HasPrefix(constraint, ">=") {
		min, err := Parse(constraint[2:])
		if err != nil {
			return false, err
		}
		return v.GTE(min), nil
	}

	// <= x.y.z
	if strings.HasPrefix(constraint, "<=") {
		max, err := Parse(constraint[2:])
		if err != nil {
			return false, err
		}
		return v.LTE(max), nil
	}

	// > x.y.z
	if strings.HasPrefix(constraint, ">") {
		min, err := Parse(constraint[1:])
		if err != nil {
			return false, err
		}
		return v.GT(min), nil
	}

	// < x.y.z
	if strings.HasPrefix(constraint, "<") {
		max, err := Parse(constraint[1:])
		if err != nil {
			return false, err
		}
		return v.LT(max), nil
	}

	// ^x.y.z - 兼容版本（不改变主版本号）
	if strings.HasPrefix(constraint, "^") {
		base, err := Parse(constraint[1:])
		if err != nil {
			return false, err
		}
		// ^1.2.3 表示 >=1.2.3 <2.0.0
		upper := &Version{Major: base.Major + 1, Minor: 0, Patch: 0}
		return v.GTE(base) && v.LT(upper), nil
	}

	// ~x.y.z - 近似版本（不改变次版本号）
	if strings.HasPrefix(constraint, "~") {
		base, err := Parse(constraint[1:])
		if err != nil {
			return false, err
		}
		// ~1.2.3 表示 >=1.2.3 <1.3.0
		upper := &Version{Major: base.Major, Minor: base.Minor + 1, Patch: 0}
		return v.GTE(base) && v.LT(upper), nil
	}

	// =x.y.z 或 x.y.z - 精确版本
	constraint = strings.TrimPrefix(constraint, "=")
	exact, err := Parse(constraint)
	if err != nil {
		return false, err
	}
	return v.EQ(exact), nil
}
