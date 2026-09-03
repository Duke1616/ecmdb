# ECMDB 全局工程规范与 AI 协作准则 (AGENTS.md)

ECMDB 是企业级下一代配置管理数据库，基于动态元数据驱动（Schema-less）、资产拓扑图引擎与弹性微服务插件体系构建。

## 1. 架构分层与依赖单向流转铁律
$$\text{Transport (Web / gRPC)} \longrightarrow \text{Service} \longrightarrow \text{Repository} \longrightarrow \text{DAO / Storage}$$
- **Domain (`internal/domain/`)**：纯 Go 领域模型，严禁依赖 Gin、GORM、MongoDB 等传输或持久化包。
- **Transport (`internal/web/`)**：薄适配层，统一使用官方 `ginx.B` 与 `Define(desc, name).Bind(...)`，**严禁编写业务逻辑，严禁直接读写 DAO**。
- **Service (`internal/service/`)**：核心元数据驱动、资产动态 Schema 映射、图拓扑分析与加密编排。禁止反向依赖 Transport。
- **Repository & DAO (`internal/repository/`)**：Repo 负责领域对象与动态文档转换；DAO 负责 MongoDB 集合、聚合管道与事务控制。
- **Plugin 中心 (`internal/web/plugin/`, `pkg/plugin/`)**：插件扩展控制面，以“插件中心”顶级域统管自描述反代与注册。

## 2. 动态元数据驱动与存储安全
- **动态集合命名**：资产数据按模型动态分表（`c_{model_uid}`），元数据统管于 `c_model`、`c_attribute`、`c_attribute_group`。
- **敏感数据异步加密**：属性安全标记变更必须投递 `field_secure_attr_change_events`，由 `ResilientConsumer` 异步批量重加密。
- **模型删除级联阻断**：删除模型前强制执行 `IDeleteModelDependencyChecker`，存量资产或拓扑关系未清理时严格阻断。

## 3. 权限契约与 EIAM 自发现规范
- **声明式路由定义**：Handler 统一使用 `h.Define("名称", "编码").Bind(...)` 声明受控端点。
- **强类型权限依赖 (Needs)**：权限依赖**强制引用 `permission.Xxx.Yyy` 强类型常量**，杜绝手写字面量字符串。
- **自动化契约生成**：路由或权限变更后，必须执行 `task perm` 生成 Go 契约代码及 `docs/permissions.md`。

## 4. 现代 Go 编码习惯与质量要求
- **集合处理规范**：切片变换、过滤、查找、去重**优先使用 `github.com/samber/lo`**，杜绝多层冗余 for 循环。
- **命名规范**：核心接口以 `I` 开头（如 `IUserRepository`），实现结构体小写不导出，构造函数为 `NewService`。
- **语言与注释**：所有交互、文档与代码注释**必须使用简体中文**；注释重点解释**“为什么这样设计”**，禁止复述字面代码。
- **错误包装**：跨层错误必须使用 `fmt.Errorf("...: %w", err)` 保留根因链。

## 5. 构建工具链与测试约束（强制）
- **自动化工具链优先**：
  - 代码与文档全量生成：`task gen`（串联 wire + proto + perm）
  - 依赖注入更新：修改 `ioc/` 后执行 `task wire`（`wire ./ioc/`）
  - 依赖整理：`go mod tidy`
- **单元测试规范（强约束）**：
  - Service / Repo 单元测试统一采用 **Table-Driven（表驱动测试）**。
  - **就近 Mock 原则**：Mock 代码统一使用 `mockgen` 生成并收敛在对应业务包的 `mocks/` 子目录中（如 `service/attribute/mocks`），严禁手写伪桩代码（Stub）。
  - 代码提交前必须确保 `go test ./...` 100% 通过。
