# ECMDB 全局权限大盘与元数据字典

> 本文档由 `permgen` 基于全仓 AST 静态分析自动生成。请勿手动修改。
>
> 💡 **联动包含机制**：当为角色分配某项操作权限时，系统将**自动附带拥有**其“联动包含”中的权限，无需管理员手动重复勾选（例如：勾选“修改用户”会自动附带拥有“用户详情”权限）。

- **受控业务模块数**: 7
- **受控权限点总数**: 62


## 模块: 模型管理/属性管理 (`attribute`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/attribute/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 创建属性 | `cmdb:attribute:add` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 81 |
| 删除属性 | `cmdb:attribute:delete` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 103 |
| 更新属性 | `cmdb:attribute:edit` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 108 |
| 创建分组 | `cmdb:attribute:group_add` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 49 |
| 删除分组 | `cmdb:attribute:group_delete` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 60 |
| 重命名分组 | `cmdb:attribute:group_rename` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 65 |
| 分组排序 | `cmdb:attribute:group_sort` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/attribute/handler.go` 行 70 |
| 批量查询分组 | `cmdb:attribute:group_view_by_ids` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/attribute/handler.go` 行 54 |
| 属性排序 | `cmdb:attribute:sort` | 租户级 | 本级 | 正常 | 分组排序 · `cmdb:attribute:group_sort` | `internal/web/attribute/handler.go` 行 113 |
| 属性列表 | `cmdb:attribute:view` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/attribute/handler.go` 行 86 |
| 自定义列展示 | `cmdb:attribute:view_custom_fields` | 租户级 | 本级 | 正常 | - | `internal/web/attribute/handler.go` 行 98 |
| 属性字段 | `cmdb:attribute:view_fields` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/attribute/handler.go` 行 92 |

---


## 模块: 资产仓库/导入导出 (`dataio`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/dataio/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 数据导出 | `cmdb:dataio:export` | 租户级 | 本级 | 正常 | - | `internal/web/dataio/handler.go` 行 44 |
| 模板导出 | `cmdb:dataio:export_template` | 租户级 | 本级 | 正常 | - | `internal/web/dataio/handler.go` 行 36 |
| 数据导入 | `cmdb:dataio:import` | 租户级 | 本级 | 正常 | - | `internal/web/dataio/handler.go` 行 40 |

---


## 模块: 模型管理 (`model`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/model/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 创建模型 | `cmdb:model:add` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 73 |
| 删除模型 | `cmdb:model:delete` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 84 |
| 模型详情 | `cmdb:model:get` | 租户级 | 本级 | 正常 | 属性列表 · `cmdb:attribute:view`<br>关联类型列表 · `cmdb:relation:view`<br>模型关联列表 · `cmdb:model:relation_view` | `internal/web/model/handler.go` 行 78 |
| 创建分组 | `cmdb:model:group_add` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 54 |
| 删除分组 | `cmdb:model:group_delete` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 59 |
| 重命名分组 | `cmdb:model:group_rename` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 64 |
| 创建模型关联关系 | `cmdb:model:relation_add` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 110 |
| 删除模型关联关系 | `cmdb:model:relation_delete` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 120 |
| 更新模型关联关系 | `cmdb:model:relation_edit` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 125 |
| 模型拓扑图 | `cmdb:model:relation_graph` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 105 |
| 模型关联列表 | `cmdb:model:relation_view` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 115 |
| 模型列表 | `cmdb:model:view` | 租户级 | 本级 | 正常 | - | `internal/web/model/handler.go` 行 89 |
| 按UID批量查询模型 | `cmdb:model:view_by_uids` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/model/handler.go` 行 94 |

---


## 模块: 插件中心/插件管理 (`plugin`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/plugin/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 查询资源插件动作 | `cmdb:plugin:actions` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/plugin/handler.go` 行 69 |
| 保存插件绑定 | `cmdb:plugin:create` | 租户级 | 本级 | 正常 | 查询默认插件定义 · `cmdb:plugin:default`<br>属性列表 · `cmdb:attribute:view`<br>模型列表 · `cmdb:model:view` | `internal/web/plugin/handler.go` 行 59 |
| 查询默认插件定义 | `cmdb:plugin:default` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/plugin/handler.go` 行 55 |
| 删除插件绑定 | `cmdb:plugin:delete` | 租户级 | 本级 | 正常 | - | `internal/web/plugin/handler.go` 行 66 |
| 插件枚举 | `cmdb:plugin:enums` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/plugin/handler.go` 行 51 |
| 插件详情 | `cmdb:plugin:get` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/plugin/handler.go` 行 47 |
| 解析插件动作 | `cmdb:plugin:resolve` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/plugin/handler.go` 行 73 |
| 插件运行时视图 | `cmdb:plugin:runtime_view` | 租户级 | 本级 | 正常 | 解析插件动作 · `cmdb:plugin:resolve` | `internal/web/plugin/handler.go` 行 77 |
| 切换插件绑定状态 | `cmdb:plugin:switch` | 租户级 | 本级 | 正常 | - | `internal/web/plugin/handler.go` 行 63 |
| 插件目录 | `cmdb:plugin:view` | 租户级 | 本级 | 正常 | 插件枚举 · `cmdb:plugin:enums`<br>插件详情 · `cmdb:plugin:get` | `internal/web/plugin/handler.go` 行 43 |

---


## 模块: 关联类型 (`relation`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/relation/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 创建关联类型 | `cmdb:relation:add` | 租户级 | 本级 | 正常 | - | `internal/web/relation/handler.go` 行 41 |
| 删除关联类型 | `cmdb:relation:delete` | 租户级 | 本级 | 正常 | - | `internal/web/relation/handler.go` 行 56 |
| 更新关联类型 | `cmdb:relation:edit` | 租户级 | 本级 | 正常 | - | `internal/web/relation/handler.go` 行 51 |
| 关联类型列表 | `cmdb:relation:view` | 租户级 | 本级 | 正常 | - | `internal/web/relation/handler.go` 行 46 |

---


## 模块: 资产仓库 (`resource`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/resource/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 创建资产 | `cmdb:resource:add` | 租户级 | 本级 | 正常 | 获取上传预签名 · `cmdb:tools:put_presigned_url` | `internal/web/resource/handler.go` 行 53 |
| 拓扑图向左拓展 | `cmdb:resource:add_relation_left` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/resource/handler.go` 行 114 |
| 拓扑图向右拓展 | `cmdb:resource:add_relation_right` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/resource/handler.go` 行 120 |
| 删除资产 | `cmdb:resource:delete` | 租户级 | 本级 | 正常 | - | `internal/web/resource/handler.go` 行 70 |
| 修改资产 | `cmdb:resource:edit` | 租户级 | 本级 | 正常 | 获取上传预签名 · `cmdb:tools:put_presigned_url` | `internal/web/resource/handler.go` 行 75 |
| 设置自定义属性 | `cmdb:resource:edit_custom_field` | 租户级 | 本级 | 正常 | - | `internal/web/resource/handler.go` 行 81 |
| 资产详情 | `cmdb:resource:get` | 租户级 | 本级 | 正常 | - | `internal/web/resource/handler.go` 行 59 |
| 查询加密字段 | `cmdb:resource:get_secure` | 租户级 | 本级 | 正常 | - | `internal/web/resource/handler.go` 行 92 |
| 创建资产关系 | `cmdb:resource:relation_add` | 租户级 | 本级 | 正常 | 查询可关联的资产列表 · `cmdb:resource:view_can_be_related` | `internal/web/resource/handler.go` 行 126 |
| 删除资产关系 | `cmdb:resource:relation_delete` | 租户级 | 本级 | 正常 | - | `internal/web/resource/handler.go` 行 139 |
| 全文检索资产 | `cmdb:resource:search` | 租户级 | 本级 | 正常 | - | `internal/web/resource/handler.go` 行 149 |
| 资产列表 | `cmdb:resource:view` | 租户级 | 本级 | 正常 | 模型列表 · `cmdb:model:view`<br>属性列表 · `cmdb:attribute:view`<br>获取下载预签名 · `cmdb:tools:get_presigned_url`<br>查询资源插件动作 · `cmdb:plugin:actions` | `internal/web/resource/handler.go` 行 64 |
| 批量查询资产 | `cmdb:resource:view_by_ids` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/resource/handler.go` 行 86 |
| 查询可关联的资产列表 | `cmdb:resource:view_can_be_related` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/resource/handler.go` 行 102 |
| 所有资产关系聚合查询 | `cmdb:resource:view_relation_all` | 租户级 | 本级 | 正常 | 关联类型列表 · `cmdb:relation:view`<br>模型关联列表 · `cmdb:model:relation_view`<br>按UID批量查询模型 · `cmdb:model:view_by_uids`<br>属性字段 · `cmdb:attribute:view_fields`<br>批量查询资产 · `cmdb:resource:view_by_ids` | `internal/web/resource/handler.go` 行 132 |
| 资产关联拓扑图 | `cmdb:resource:view_relation_graph` | 租户级 | 本级 | 正常 | 拓扑图向左拓展 · `cmdb:resource:add_relation_left`<br>拓扑图向右拓展 · `cmdb:resource:add_relation_right` | `internal/web/resource/handler.go` 行 108 |

---


## 模块: 资产仓库/文件管理 (`tools`)

- **所属服务**: `cmdb`
- **定义源码**: `internal/web/tools/handler.go`

| 操作名称 | 完整权限码 | 作用域 | 归属类型 | 暴露状态 | 联动包含权限 | 宿主源码位置 |
|:---|:---|:---|:---|:---|:---|:---|
| 获取下载预签名 | `cmdb:tools:get_presigned_url` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/tools/handler.go` 行 44 |
| 获取上传预签名 | `cmdb:tools:put_presigned_url` | 租户级 | 本级 | 静默 (不暴露) | - | `internal/web/tools/handler.go` 行 48 |
| 删除对象 | `cmdb:tools:remove` | 租户级 | 本级 | 正常 | - | `internal/web/tools/handler.go` 行 52 |
| 文件上传 | `cmdb:tools:upload` | 租户级 | 本级 | 正常 | - | `internal/web/tools/handler.go` 行 40 |

---


