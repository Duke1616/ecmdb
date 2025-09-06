# 菜单数据 AST 代码生成器

这个工具使用 Go AST 技术根据 JSON 文件生成菜单相关的 Go 代码。

## 功能特性

- ✅ 使用 Go AST 技术生成类型安全的 Go 代码
- ✅ 自动解析 JSON 菜单数据
- ✅ 生成符合项目结构的 `domain.Menu` 类型
- ✅ 支持菜单层级结构和权限端点
- ✅ 集成到系统初始化流程中

## 文件结构

```
cmd/
├── initial/menu/
│   ├── menu_data.go        # 生成的菜单数据文件
│   └── README.md           # 本说明文档
└── tools/menu-generator/
    └── ast_generator.go    # AST 代码生成器
```

## 使用方法

### 1. 更新菜单数据

编辑 `init/c_menu.json` 文件来修改菜单数据：

```json
[
  {
    "id": "1",
    "pid": "0",
    "path": "/dashboard",
    "name": "dashboard",
    "sort": "1",
    "component": "/views/dashboard/index.vue",
    "redirect": "",
    "status": "1",
    "type": "2",
    "meta": "{\"title\": \"仪表盘\", \"is_hidden\": false, \"is_affix\": true, \"is_keepalive\": true, \"icon\": \"dashboard\", \"platform\": \"cmdb\"}",
    "endpoints": "[]"
  }
]
```

### 2. 生成代码

运行脚本自动生成代码：

```bash
./scripts/update-menu.sh
```

或者手动生成：

```bash
cd cmd/initial/menu
go run ast_generator.go ../../../init/c_menu.json
```

### 3. 初始化系统

运行系统初始化：

```bash
./ecmdb init
```

## 生成的文件

生成的 `menu_data.go` 文件包含：

- `DefaultMenus` 变量：包含所有菜单数据的切片
- `GetDefaultMenus()` 函数：返回菜单数据

## 技术实现

### AST 生成器特性

1. **类型安全**：使用 Go AST 生成符合 `domain.Menu` 结构的代码
2. **自动解析**：自动解析 JSON 中的 `meta` 和 `endpoints` 字段
3. **格式化输出**：使用 `go/format` 包确保生成的代码格式正确
4. **错误处理**：完善的错误处理和验证机制

### 支持的数据类型

- **菜单类型**：目录 (1)、菜单 (2)、按钮 (3)
- **状态**：启用 (1)、禁用 (2)
- **元数据**：标题、图标、平台、隐藏状态等
- **端点**：API 路径、方法、描述

## 注意事项

1. 确保 JSON 文件格式正确
2. 生成的代码会自动覆盖 `menu_data.go` 文件
3. 修改菜单数据后需要重新生成代码
4. 系统初始化时会检查是否已有菜单数据，避免重复创建

## 故障排除

如果遇到问题：

1. 检查 JSON 文件格式是否正确
2. 确保所有必需的字段都存在
3. 查看编译错误信息
4. 检查数据库连接是否正常
