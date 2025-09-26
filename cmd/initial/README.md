# Initial 模块使用说明

## 概述

Initial 模块提供了完整的数据库初始化和版本管理功能，支持全量初始化、增量更新、版本回滚等操作。

## 功能特性

- ✅ 全量数据初始化
- ✅ 增量版本更新
- ✅ 指定版本执行
- ✅ 版本回滚功能
- ✅ 版本列表查看
- ✅ 干运行模式预览
- ✅ 版本状态管理

## 命令使用

### 1. 基本初始化

```bash
# 全量初始化（首次运行）
go run main.go init

# 增量更新到最新版本
go run main.go init
```

### 2. 指定版本执行

```bash
# 更新到指定版本
go run main.go init --version v1.2.3

# 或者使用短参数
go run main.go init -v v1.2.3
```

### 3. 干运行模式

```bash
# 预览操作，不执行实际更新
go run main.go init --dry-run

# 预览指定版本的操作
go run main.go init --version v1.2.3 --dry-run
```

### 4. 版本回滚

```bash
# 回滚到指定版本
go run main.go init rollback v1.2.3

# 预览回滚操作
go run main.go init rollback v1.2.3 --dry-run
```

### 5. 版本列表

```bash
# 查看所有可用版本
go run main.go init list
```

## 版本管理

### 版本格式

版本号遵循语义化版本规范：`v主版本.次版本.修订版本`

例如：`v1.2.3`

### 版本状态

- **未执行**: 版本尚未执行
- **已执行**: 版本已执行完成
- **当前版本**: 系统当前运行的版本

### 版本比较规则

版本号按数字大小进行比较，例如：
- `v1.2.3` < `v1.2.4`
- `v1.2.3` < `v1.3.0`
- `v1.2.3` < `v2.0.0`

## 增量更新流程

每个版本更新包含以下步骤：

1. **Before**: 更新前的准备工作
2. **Commit**: 执行实际的更新操作
3. **After**: 更新后的清理工作

## 回滚流程

回滚操作会按版本号降序执行每个版本的 `Rollback` 方法。

## 开发新版本

要添加新的版本更新，需要：

1. 在 `cmd/initial/incr/` 目录下创建新的版本文件
2. 实现 `InitialIncr` 接口
3. 在 `RegisterIncr` 函数中注册新版本

### 示例版本文件

```go
package incr

import (
	"context"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
)

type incrV200 struct {
	App *ioc.App
}

func NewIncrV200(app *ioc.App) InitialIncr {
	return &incrV200{App: app}
}

func (i *incrV200) Version() string {
	return "v2.0.0"
}

func (i *incrV200) Before(ctx context.Context) error {
	// 更新前的准备工作
	return nil
}

func (i *incrV200) Commit(ctx context.Context) error {
	// 执行实际的更新操作
	return nil
}

func (i *incrV200) After(ctx context.Context) error {
	// 更新后的清理工作
	return i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version())
}

func (i *incrV200) Rollback(ctx context.Context) error {
	// 回滚操作
	return nil
}
```

## 注意事项

1. **版本顺序**: 版本号必须按顺序递增，不能跳跃
2. **回滚安全**: 回滚操作会删除数据，请谨慎使用
3. **事务处理**: 建议在版本更新中使用数据库事务
4. **错误处理**: 更新失败时会停止后续操作
5. **备份建议**: 重要操作前建议备份数据库

## 故障排除

### 常见问题

1. **版本重复错误**: 检查版本号是否已存在
2. **版本顺序错误**: 确保版本号按顺序递增
3. **回滚失败**: 检查回滚方法实现是否正确
4. **权限问题**: 确保有足够的数据库操作权限

### 调试模式

```bash
# 启用调试模式查看详细信息
go run main.go init --debug
```
