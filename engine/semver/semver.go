// Package semver 提供语义化版本处理功能。
package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 预定义错误变量
var (
	ErrInvalidVersion = errors.New("invalid version format")
	ErrParseVersion1  = errors.New("failed to parse version 1")
	ErrParseVersion2  = errors.New("failed to parse version 2")
)

// Version 表示语义化版本。
// 字段按内存对齐优化排序
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
		return nil, fmt.Errorf("%w: %s", ErrInvalidVersion, version)
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
		return 0, fmt.Errorf("%w: %w", ErrParseVersion1, err)
	}
	ver2, err := Parse(v2)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrParseVersion2, err)
	}
	return ver1.Compare(ver2), nil
}

// IsValid 检查版本字符串是否有效。
func IsValid(version string) bool {
	_, err := Parse(version)
	return err == nil
}

// ConstraintType 表示约束类型
type ConstraintType int

const (
	ConstraintGTE ConstraintType = iota // >=
	ConstraintLTE                       // <=
	ConstraintGT                        // >
	ConstraintLT                        // <
	ConstraintCaret                     // ^
	ConstraintTilde                     // ~
	ConstraintExact                     // = or no prefix
)

// parseConstraint 解析约束字符串
func parseConstraint(constraint string) (ConstraintType, string, error) {
	constraint = strings.TrimSpace(constraint)

	switch {
	case strings.HasPrefix(constraint, ">="):
		return ConstraintGTE, strings.TrimSpace(constraint[2:]), nil
	case strings.HasPrefix(constraint, "<="):
		return ConstraintLTE, strings.TrimSpace(constraint[2:]), nil
	case strings.HasPrefix(constraint, ">"):
		return ConstraintGT, strings.TrimSpace(constraint[1:]), nil
	case strings.HasPrefix(constraint, "<"):
		return ConstraintLT, strings.TrimSpace(constraint[1:]), nil
	case strings.HasPrefix(constraint, "^"):
		return ConstraintCaret, strings.TrimSpace(constraint[1:]), nil
	case strings.HasPrefix(constraint, "~"):
		return ConstraintTilde, strings.TrimSpace(constraint[1:]), nil
	default:
		constraint = strings.TrimPrefix(constraint, "=")
		return ConstraintExact, strings.TrimSpace(constraint), nil
	}
}

// Satisfies 检查版本是否满足约束。
func Satisfies(version, constraint string) (bool, error) {
	v, err := Parse(version)
	if err != nil {
		return false, err
	}

	conType, conVersion, err := parseConstraint(constraint)
	if err != nil {
		return false, err
	}

	switch conType {
	case ConstraintGTE:
		min, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		return v.GTE(min), nil
	case ConstraintLTE:
		max, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		return v.LTE(max), nil
	case ConstraintGT:
		min, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		return v.GT(min), nil
	case ConstraintLT:
		max, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		return v.LT(max), nil
	case ConstraintCaret:
		base, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		// ^1.2.3 表示 >=1.2.3 <2.0.0
		upper := &Version{Major: base.Major + 1, Minor: 0, Patch: 0}
		return v.GTE(base) && v.LT(upper), nil
	case ConstraintTilde:
		base, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		// ~1.2.3 表示 >=1.2.3 <1.3.0
		upper := &Version{Major: base.Major, Minor: base.Minor + 1, Patch: 0}
		return v.GTE(base) && v.LT(upper), nil
	case ConstraintExact:
		exact, err := Parse(conVersion)
		if err != nil {
			return false, err
		}
		return v.EQ(exact), nil
	default:
		return false, fmt.Errorf("%w: %s", ErrInvalidVersion, constraint)
	}
}
