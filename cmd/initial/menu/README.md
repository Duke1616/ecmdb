# 菜单管理模块

本模块提供了菜单的初始化、更新和 MD5 检查功能。

## 功能特性

- **MD5 哈希检查**: 通过计算菜单数据的 MD5 哈希值来判断菜单是否需要更新
- **智能更新**: 只有当菜单数据发生变化时才执行菜单更新操作
- **版本管理**: 支持将菜单哈希值存储到版本信息中
- **权限同步**: 菜单更新后自动同步相关权限
- **编译兼容**: 基于菜单数据内容计算哈希，确保在编译后也能正常工作

## 文件结构

```
cmd/initial/menu/
├── menu_data.go      # 菜单数据定义
├── menu_service.go   # 菜单服务实现
├── hash.go          # MD5 哈希计算工具
├── example.go       # 使用示例
└── README.md        # 说明文档
```

## 使用方法

### 1. 全量初始化 (Full Initialization)

在首次安装时，直接插入菜单数据并记录 MD5：

```go
// 在 full 初始化中使用
func (i *fullInitial) InitMenu() error {
    // ... 菜单初始化逻辑 ...
    
    // 计算并存储菜单文件的 MD5 哈希值
    hashCalculator := menu.NewMenuHashCalculator()
    menuHash, err := hashCalculator.CalculateProjectMenuHash()
    if err == nil {
        i.App.VerSvc.SetMenuHash(ctx, "v1.0.0", menuHash)
    }
    
    return nil
}
```

### 2. 增量更新 (Incremental Update)

在版本更新时，先检查 MD5 再决定是否更新：

```go
// 在版本更新中使用
func (i *incrV192) Commit(ctx context.Context) error {
    // 创建菜单服务
    menuService := menu.NewMenu(i.App)
    
    // 使用智能菜单更新（带 MD5 检查）
    if err := menuService.UpdateMenu(i.App); err != nil {
        return err
    }
    
    return nil
}
```

### 3. 手动计算菜单哈希

```go
hashCalculator := menu.NewMenuHashCalculator()

// 计算菜单数据的哈希值
hash, err := hashCalculator.CalculateMenuHash()
if err != nil {
    log.Printf("计算菜单哈希失败: %v", err)
    return
}

log.Printf("菜单哈希值: %s", hash)
```

## API 参考

### MenuHashCalculator

菜单哈希计算器，用于计算菜单文件的 MD5 哈希值。

#### 方法

- `CalculateMenuHash() (string, error)`: 计算菜单数据的哈希值
- `CalculateProjectMenuHash() (string, error)`: 计算项目菜单的整体哈希值（与 CalculateMenuHash 相同）
- `CalculateMenuDataHash(menus interface{}) (string, error)`: 计算指定菜单数据的哈希值

### Menu

菜单服务，提供菜单更新和权限同步功能。

#### 方法

- `ChangeMenu(app *ioc.App)`: 传统的菜单更新方法
- `UpdateMenu(app *ioc.App) error`: 智能菜单更新方法（带 MD5 检查）

## 配置说明

### 菜单数据源

哈希计算基于菜单数据内容：
- 使用 `GetInjectMenus()` 获取当前菜单数据
- 通过 JSON 序列化确保稳定性和一致性
- 不依赖文件系统，确保在编译后也能正常工作

### 版本存储

菜单哈希值存储在 MongoDB 的 `c_version` 集合中，字段为 `menu_hash`。菜单哈希与当前版本关联，而不是特定版本，这样设计更合理：

- **菜单哈希**: 反映当前菜单数据的状态
- **版本信息**: 记录当前系统的版本
- **关联关系**: 菜单哈希存储在版本记录中，表示该版本对应的菜单状态

## 注意事项

1. **数据监控**: 系统会监控菜单数据内容的变化，基于实际菜单数据计算哈希
2. **哈希一致性**: 确保所有环境使用相同的菜单数据，避免哈希值不一致
3. **权限同步**: 菜单更新后会自动同步相关权限，无需手动处理
4. **错误处理**: 如果哈希计算失败，系统会继续执行菜单更新操作
5. **编译兼容**: 基于菜单数据内容计算哈希，确保在编译后的二进制文件中也能正常工作

## 示例场景

### 场景1: 菜单文件未变化

```
[INFO] 开始检查菜单是否需要更新 版本=v1.9.2
[INFO] 当前菜单文件哈希 hash=abc123...
[INFO] 存储的菜单哈希 hash=abc123...
[INFO] 菜单文件未发生变化，跳过菜单更新
```

### 场景2: 菜单文件已变化

```
[INFO] 开始检查菜单是否需要更新 版本=v1.9.2
[INFO] 当前菜单文件哈希 hash=def456...
[INFO] 存储的菜单哈希 hash=abc123...
[INFO] 菜单文件已发生变化，开始更新菜单
[INFO] 菜单数据已更新 修改数量=5 插入数量=2 删除数量=1
[INFO] 菜单更新完成 hash=def456...
```