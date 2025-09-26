# 版本增量更新目录结构

## 目录说明

每个版本都有独立的目录，包含以下文件：

```
incr/
├── types.go                    # 版本管理核心逻辑
├── version_test.go            # 版本比较测试
├── README.md                  # 本说明文档
├── v1.5.0/                    # v1.5.0 版本目录
│   ├── incr-v1.5.0.go        # 版本实现
│   ├── incr_v150_test.go     # 版本测试
│   └── CHANGES.md            # 版本变更说明
└── v1.9.2/                    # v1.9.2 版本目录
    ├── incr-v1.9.2.go        # 版本实现
    ├── incr_v192_test.go     # 版本测试
    └── CHANGES.md            # 版本变更说明
```

## 版本目录结构

每个版本目录应包含：

### 必需文件
- `incr-v{version}.go` - 版本实现文件
- `incr_v{version}_test.go` - 版本测试文件
- `CHANGES.md` - 版本变更说明文档

### 可选文件
- `README.md` - 版本特殊说明
- `migration.sql` - 数据库迁移脚本
- `rollback.sql` - 回滚脚本
- `config.yaml` - 版本配置

## 创建新版本

### 1. 创建版本目录
```bash
mkdir -p cmd/initial/incr/v{version}
```

### 2. 创建实现文件
```go
// incr-v{version}.go
package v{version}

import (
    "context"
    "github.com/Duke1616/ecmdb/cmd/initial/ioc"
    "github.com/Duke1616/ecmdb/cmd/initial/incr"
)

type incrV{version} struct {
    App *ioc.App
}

func NewIncrV{version}(app *ioc.App) incr.InitialIncr {
    return &incrV{version}{
        App: app,
    }
}

func (i *incrV{version}) Version() string {
    return "v{version}"
}

func (i *incrV{version}) Commit(ctx context.Context) error {
    // 实现版本更新逻辑
    return nil
}

func (i *incrV{version}) Rollback(ctx context.Context) error {
    // 实现版本回滚逻辑
    return nil
}

func (i *incrV{version}) Before(ctx context.Context) error {
    // 实现更新前逻辑（如备份）
    return nil
}

func (i *incrV{version}) After(ctx context.Context) error {
    // 实现更新后逻辑（如更新版本号）
    return i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version())
}
```

### 3. 创建测试文件
```go
// incr_v{version}_test.go
package v{version}

import (
    "testing"
)

func TestVersion{version}Logic(t *testing.T) {
    // 实现版本测试逻辑
}
```

### 4. 创建变更文档
```markdown
# v{version} 版本更新

## 更新内容
- 功能更新1
- 功能更新2

## 数据库变更
- 表结构变更
- 数据迁移

## 升级说明
- 前置条件
- 升级步骤
- 回滚说明
```

### 5. 注册版本
在 `types.go` 中添加导入和注册：

```go
import (
    // ... 其他导入
    "github.com/Duke1616/ecmdb/cmd/initial/incr/v{version}"
)

func RegisterIncr(app *ioc.App) {
    // ... 其他注册
    registerIncr(v{version}.NewIncrV{version}(app))
}
```

## 版本命名规范

- 版本目录：`v{major}.{minor}.{patch}`
- 实现文件：`incr-v{major}.{minor}.{patch}.go`
- 测试文件：`incr_v{major}{minor}{patch}_test.go`
- 包名：`v{major}{minor}{patch}`（如 v150, v192）
- 构造函数：`NewIncrV{major}{minor}{patch}`

## 版本实现接口

每个版本必须实现 `InitialIncr` 接口：

```go
type InitialIncr interface {
    Version() string                    // 返回版本号
    Rollback(ctx context.Context) error // 回滚逻辑
    Commit(ctx context.Context) error   // 更新逻辑
    Before(ctx context.Context) error   // 更新前逻辑
    After(ctx context.Context) error    // 更新后逻辑
}
```

## 最佳实践

1. **备份优先**: 在 `Before` 方法中执行数据备份
2. **事务安全**: 确保 `Commit` 和 `Rollback` 的原子性
3. **错误处理**: 完善的错误处理和日志记录
4. **测试覆盖**: 为每个版本编写完整的测试
5. **文档完整**: 详细记录版本变更和升级说明

## 版本执行顺序

版本按照语义版本号顺序执行，确保：
- 版本号较小的先执行
- 版本号较大的后执行
- 支持指定目标版本执行
- 支持回滚到指定版本
