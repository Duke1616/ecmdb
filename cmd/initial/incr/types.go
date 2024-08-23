package incr

import (
	"fmt"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/spf13/cobra"
	"regexp"
	"strconv"
)

// ErrDuplicateVersion 表示版本重复的错误
var ErrDuplicateVersion = fmt.Errorf("版本重复，请修改")

type InitialIncr interface {
	Version() string
	Rollback() error
	Commit() error
}

var incrRegistry = make(map[string]InitialIncr)

func registerIncr(incr InitialIncr) {
	if _, ok := incrRegistry[incr.Version()]; ok {
		cobra.CheckErr(ErrDuplicateVersion)
	}

	incrRegistry[incr.Version()] = incr
}

func RegisterIncr(app *ioc.App) {
	registerIncr(NewIncrV123(app))
	registerIncr(NewIncrV130(app))
}

// RunIncrementalOperations 执行所有低于当前版本的增量操作
func RunIncrementalOperations(currentVersion string) error {
	// 将当前版本号解析为一个切片，以便进行比较
	currentVerSlice := parseVersion(currentVersion)

	// 遍历注册表中的所有增量操作
	for version, incr := range incrRegistry {
		// 将注册表中的版本号解析为切片
		versionSlice := parseVersion(version)

		// 如果当前版本高于注册表中的版本，则执行该增量操作
		if less(versionSlice, currentVerSlice) {
			fmt.Printf("Executing incremental operation for version %s...\n", version)
			err := incr.Commit()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// parseVersion 将版本号字符串转换为整数切片，以便进行比较
func parseVersion(version string) []int {
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
func less(v1, v2 []int) bool {
	for i := range v1 {
		if v1[i] > v2[i] {
			return true
		} else if v1[i] < v2[i] {
			return false
		}
	}
	return false
}
