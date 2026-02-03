package register

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/incr/v1.5.0"
	"github.com/Duke1616/ecmdb/cmd/initial/incr/v1.9.2"
	v193 "github.com/Duke1616/ecmdb/cmd/initial/incr/v1.9.3"
	v194 "github.com/Duke1616/ecmdb/cmd/initial/incr/v1.9.4"
	"github.com/Duke1616/ecmdb/cmd/initial/incr/version"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/spf13/cobra"
)

// ErrDuplicateVersion 表示版本重复的错误
var ErrDuplicateVersion = fmt.Errorf("版本重复，请修改")

var incrRegistry = make(map[string]incr.InitialIncr)

func registerIncr(incr incr.InitialIncr) {
	if _, ok := incrRegistry[incr.Version()]; ok {
		cobra.CheckErr(ErrDuplicateVersion)
	}

	incrRegistry[incr.Version()] = incr
}

// RegisterIncr 依照版本顺序进行注册
func RegisterIncr(app *ioc.App) {
	registerIncr(v150.NewIncrV150(app))
	registerIncr(v192.NewIncrV192(app))
	registerIncr(v193.NewIncrV193(app))
	registerIncr(v194.NewIncrV194(app))
}

// RunIncrementalOperationsToVersion 执行到指定版本的增量操作
func RunIncrementalOperationsToVersion(currentVersion, targetVersion string) error {
	// 创建外层 context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	currentVerSlice := version.ParseVersion(currentVersion)
	targetVerSlice := version.ParseVersion(targetVersion)

	// 获取所有注册表中的版本号并排序
	versions := sortedVersions()

	for _, v := range versions {
		versionSlice := version.ParseVersion(v)

		// 如果版本大于当前版本且小于等于目标版本，则执行增量操作
		if version.Compare(versionSlice, currentVerSlice) && (version.Compare(targetVerSlice, versionSlice) || version.Equal(versionSlice, targetVerSlice)) {
			fmt.Printf("✅ 正在执行版本 %s 的增量操作...\n", v)

			// 执行 Before
			if err := incrRegistry[v].Before(ctx); err != nil {
				return fmt.Errorf("版本 %s Before 执行失败: %w", v, err)
			}
			// 执行 Commit
			if err := incrRegistry[v].Commit(ctx); err != nil {
				return fmt.Errorf("版本 %s Commit 执行失败: %w", v, err)
			}

			// 执行 After
			if err := incrRegistry[v].After(ctx); err != nil {
				return fmt.Errorf("版本 %s After 执行失败: %w", v, err)
			}

			fmt.Printf("版本 %s 的增量操作完成\n", v)
		}
	}

	fmt.Println("所有增量操作执行完成")
	return nil
}

// RollbackToVersion 回滚到指定版本
func RollbackToVersion(currentVersion, targetVersion string) error {
	// 创建外层 context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	currentVerSlice := version.ParseVersion(currentVersion)
	targetVerSlice := version.ParseVersion(targetVersion)

	// 获取所有注册表中的版本号并排序（降序）
	versions := sortedVersionsDesc()

	for _, v := range versions {
		versionSlice := version.ParseVersion(v)

		// 如果版本大于目标版本且小于等于当前版本，则执行回滚操作
		if version.Compare(versionSlice, targetVerSlice) && (version.Compare(currentVerSlice, versionSlice) || version.Equal(versionSlice, currentVerSlice)) {
			fmt.Printf("正在回滚版本 %s...\n", v)

			// 执行 Rollback
			if err := incrRegistry[v].Rollback(ctx); err != nil {
				return fmt.Errorf("版本 %s 回滚失败: %w", v, err)
			}

			fmt.Printf("版本 %s 回滚完成\n", v)
		}
	}

	fmt.Println("所有回滚操作执行完成")
	return nil
}

// GetAllVersions 获取所有可用版本
func GetAllVersions() []string {
	return sortedVersions()
}

// RunIncrementalOperations 执行所有低于当前版本的增量操作
func RunIncrementalOperations(currentVersion string) error {
	// 创建外层 context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	currentVerSlice := version.ParseVersion(currentVersion)

	// 获取所有注册表中的版本号并排序
	versions := sortedVersions()

	for _, v := range versions {
		versionSlice := version.ParseVersion(v)

		// 如果注册表版本大于当前版本，则执行增量操作
		if version.Compare(versionSlice, currentVerSlice) {
			fmt.Printf("正在执行版本 %s 的增量操作...\n", v)

			// 执行 Before
			if err := incrRegistry[v].Before(ctx); err != nil {
				return fmt.Errorf("版本 %s Before 执行失败: %w", v, err)
			}

			// 执行 Commit
			if err := incrRegistry[v].Commit(ctx); err != nil {
				return fmt.Errorf("版本 %s Commit 执行失败: %w", v, err)
			}

			// 执行 After
			if err := incrRegistry[v].After(ctx); err != nil {
				return fmt.Errorf("版本 %s After 执行失败: %w", v, err)
			}

			fmt.Printf("版本 %s 的增量操作完成\n", v)
		}
	}

	fmt.Println("所有增量操作执行完成")
	return nil
}

// sortedVersions 返回排序后的版本号切片（升序）
func sortedVersions() []string {
	// 将注册表中的版本号提取到一个切片
	var versions []string
	for version := range incrRegistry {
		versions = append(versions, version)
	}

	// 使用自定义的排序规则进行排序（升序）
	sort.SliceStable(versions, func(i, j int) bool {
		v1 := version.ParseVersion(versions[i])
		v2 := version.ParseVersion(versions[j])
		return version.Compare(v2, v1)
	})

	return versions
}

// sortedVersionsDesc 返回排序后的版本号切片（降序）
func sortedVersionsDesc() []string {
	// 将注册表中的版本号提取到一个切片
	var versions []string
	for version := range incrRegistry {
		versions = append(versions, version)
	}

	// 使用自定义的排序规则进行排序（降序）
	sort.SliceStable(versions, func(i, j int) bool {
		v1 := version.ParseVersion(versions[i])
		v2 := version.ParseVersion(versions[j])
		return version.Compare(v2, v1) // 注意这里调换了 v1 和 v2 的位置
	})

	return versions
}
