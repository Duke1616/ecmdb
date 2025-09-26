# 备份命令使用说明

## 概述

备份命令提供了完整的数据库备份和恢复功能，支持 MongoDB 和 MySQL 数据库。

## 命令结构

```
ecmdb backup [子命令] [参数]
```

## 可用命令

### 1. MongoDB 备份

```bash
# 备份 MongoDB 集合
ecmdb backup mongo [集合名称] [选项]

# 示例
ecmdb backup mongo c_menu --version v1.9.2 --description "升级前备份"
```

### 2. MySQL 备份

```bash
# 备份 MySQL 表
ecmdb backup mysql [表名称] [选项]

# 示例
ecmdb backup mysql casbin_rule --version v1.9.2 --description "权限表备份"
```

### 3. 列出备份记录

```bash
# 列出备份记录
ecmdb backup list [源名称] [类型]

# 示例
ecmdb backup list c_menu mongo
ecmdb backup list casbin_rule mysql
```

### 4. 恢复 MongoDB 集合

```bash
# 恢复 MongoDB 集合
ecmdb backup restore-mongo [集合名称] [备份ID]

# 示例
ecmdb backup restore-mongo c_menu v1.9.2_c_menu_20231201_143022
```

### 5. 恢复 MySQL 表

```bash
# 恢复 MySQL 表
ecmdb backup restore-mysql [表名称] [备份ID]

# 示例
ecmdb backup restore-mysql casbin_rule v1.9.2_casbin_rule_20231201_143022
```

## 全局选项

- `--version, -v`: 备份版本号 (默认: v1.0.0)
- `--description, -d`: 备份描述

## 使用示例

### 完整备份流程

```bash
# 1. 备份 MongoDB 集合
ecmdb backup mongo c_menu --version v1.9.2 --description "升级前菜单备份"

# 2. 备份 MySQL 表
ecmdb backup mysql casbin_rule --version v1.9.2 --description "升级前权限备份"

# 3. 查看备份记录
ecmdb backup list c_menu mongo
ecmdb backup list casbin_rule mysql

# 4. 恢复数据（如需要）
ecmdb backup restore-mongo c_menu v1.9.2_c_menu_20231201_143022
ecmdb backup restore-mysql casbin_rule v1.9.2_casbin_rule_20231201_143022
```