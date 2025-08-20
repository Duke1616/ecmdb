package incr

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/spf13/cobra"
)

// ErrDuplicateVersion 表示版本重复的错误
var ErrDuplicateVersion = fmt.Errorf("版本重复，请修改")

type InitialIncr interface {
	Version() string
	Rollback() error
	Commit() error
	Before() error
	After() error
}

var incrRegistry = make(map[string]InitialIncr)

func registerIncr(incr InitialIncr) {
	if _, ok := incrRegistry[incr.Version()]; ok {
		cobra.CheckErr(ErrDuplicateVersion)
	}

	incrRegistry[incr.Version()] = incr
}

// RegisterIncr 依照版本顺序进行注册
func RegisterIncr(app *ioc.App) {
	registerIncr(NewIncrV123(app))
	registerIncr(NewIncrV130(app))
	registerIncr(NewIncrV150(app))
}

// RunIncrementalOperations 执行所有低于当前版本的增量操作
func RunIncrementalOperations(currentVersion string) error {
	// 将当前版本号解析为一个切片，以便进行比较
	currentVerSlice := parseVersion(currentVersion)

	// 获取所有版本号并进行排序
	versions := sortedVersions()

	// 遍历注册表中的所有增量操作
	for _, version := range versions {
		// 将注册表中的版本号解析为切片
		versionSlice := parseVersion(version)

		// 如果当前版本高于注册表中的版本，则执行该增量操作
		if compare(versionSlice, currentVerSlice) {
			fmt.Printf("Executing incremental operation for version %s...\n", version)
			// 程序运行前
			err := incrRegistry[version].Before()
			cobra.CheckErr(err)

			// 执行提交操作
			err = incrRegistry[version].Commit()
			cobra.CheckErr(err)

			// 程序运行后
			err = incrRegistry[version].After()
			cobra.CheckErr(err)
		}
	}

	return nil
}

// sortedVersions 返回排序后的版本号切片
func sortedVersions() []string {
	// 将注册表中的版本号提取到一个切片
	var versions []string
	for version := range incrRegistry {
		versions = append(versions, version)
	}

	// 使用自定义的排序规则进行排序
	sort.SliceStable(versions, func(i, j int) bool {
		v1 := parseVersion(versions[i])
		v2 := parseVersion(versions[j])
		return compare(v1, v2)
	})

	return versions
}

// parseVersion 将版本号字符串转换为整数切片，以便进行比较
func parseVersion(version string) []int {
	// 去除版本号前的"v"字
	version = strings.TrimPrefix(version, "v")

	parts := splitVersion(version)
	versionSlice := make([]int, len(parts))
	for i, part := range parts {
		versionSlice[i], _ = strconv.Atoi(part)
	}
	return versionSlice
}

// splitVersion 将版本号字符串按照点分割成切片
func splitVersion(version string) []string {
	return regexp.MustCompile(`\.`).Split(version, -1)
}

// less 比较两个版本号的大小
func compare(v1, v2 []int) bool {
	for i := range v1 {
		if v1[i] > v2[i] {
			return true
		} else if v1[i] < v2[i] {
			return false
		}
	}
	return false
}
