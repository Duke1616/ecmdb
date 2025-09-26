package version

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseVersion 将版本号字符串转换为整数切片，以便进行比较
func ParseVersion(version string) []int {
	// 去除版本号前的"v"字
	version = strings.TrimPrefix(version, "v")

	parts := splitVersion(version)
	versionSlice := make([]int, len(parts))

	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			// 如果解析失败，返回空切片
			return []int{}
		}
		versionSlice[i] = num
	}

	return versionSlice
}

// splitVersion 将版本号字符串按点分割
func splitVersion(version string) []string {
	return regexp.MustCompile(`\.`).Split(version, -1)
}

// Compare 比较两个版本号的大小，v1 > v2 返回 true
func Compare(v1, v2 []int) bool {
	// 确保两个版本号长度一致，不足的补0
	maxLen := len(v1)
	if len(v2) > maxLen {
		maxLen = len(v2)
	}

	// 补齐长度
	for len(v1) < maxLen {
		v1 = append(v1, 0)
	}
	for len(v2) < maxLen {
		v2 = append(v2, 0)
	}

	for i := range v1 {
		if v1[i] > v2[i] {
			return true
		} else if v1[i] < v2[i] {
			return false
		}
	}
	return false
}

// Equal 比较两个版本号是否相等
func Equal(v1, v2 []int) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			return false
		}
	}
	return true
}
