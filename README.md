# CMDB + 工单一体化平台
## 前言
这个系统最初设计是为了业余学习以及结合工作需求开发的，目标是创建一个CMDB资产管理系统。完成初步开发后，
发现仅仅记录资产信息并不能完全体现平台的价值，且不支持自动化录入显得功能过于单薄。

经过一段时间的思考后，决定将项目方向调整为工单系统。借助工单系统的能力，实现数据的自动化录入到CMDB中。起初考虑重新建立一个仓库来开发这个系统，但由于人手和精力有限，最终选择在原有仓库中进行编写和完善。

最后希望能找到志同道合的伙伴，一起参与到这个项目，共同协同开发，如果有感兴趣的，可以联系我。

wx: lkz-1008

## 项目现状
当前系统的 CMDB 和工单功能已进入 GA（General Availability）阶段，欢迎大家使用！如果有需求或建议，欢迎提交 issue。请为我们点个 Star 支持一下吧！
## 项目部署
系统默认账户：admin  系统默认密码：123456
### docker
```shell
# 创建一个新的虚拟网络
docker network create ecmdb

# 启动服务
docker compose -p ecmdb -f deploy/docker-compose.yaml up -d

# 创建用户
curl -L 'http://127.0.0.1:8666/api/user/register' \
-H 'Content-Type: application/json' \
-d '{
    "username": "admin",
    "password": "123456",
    "re_password": "123456",
    "display_name": "系统管理员"
}'

# 同步权限数据
docker cp ./init/menu.tar.gz ecmdb-mongo:/mnt
docker cp ./init/role.tar.gz ecmdb-mongo:/mnt
docker exec ecmdb-mongo mongorestore --uri="mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --gzip  --collection c_menu --archive=/mnt/menu.tar.gz
docker exec ecmdb-mongo mongorestore --uri="mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --gzip  --collection c_role --archive=/mnt/role.tar.gz

# 修正 ID 自增值
docker exec ecmdb-mongo mongosh "mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --eval 'db.c_id_generator.insertOne({ name: "c_role", next_id: NumberLong("5") })'
docker exec ecmdb-mongo mongosh "mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --eval 'db.c_id_generator.insertOne({ name: "c_menu", next_id:  NumberLong("163") })'

# 导入 Casbin 权限数据
docker exec -i ecmdb-mysql mysql -uecmdb -p123456 ecmdb < ./init/casbin_rule.sql

# 用户添加权限
docker exec ecmdb-mongo mongosh "mongodb://ecmdb:123456@127.0.0.1:27017/ecmdb?authSource=admin" --eval 'db.c_user.updateOne( { username: "admin" }, { $set: { role_codes: ["admin"] } } )'

# 重启后端服务，加载策略
docker restart ecmdb

# 环境销毁
docker compose -p ecmdb -f deploy/docker-compose.yaml down
```

## 关联项目
具体内容请查看相关项目，如果没有工单自动化任务需求，可以不部署任务执行器
- 前端： https://github.com/Duke1616/ecmdb-web
- 任务执行器：https://github.com/Duke1616/ework-runner
- 消息通知：https://github.com/Duke1616/enotify

## 功能
- CMDB
  - 全文检索、针对资产数据全文检索
  - 提供了模型的抽象管理，自定义模型属性、加密属性
  - 模型关联关系、资产关联关系
  - 集成专用登录网关资产，支持 Web SSH 和 Web SFTP 功能
- 工单中心
  - 前端集成`form-create`可以自定义工单模版
  - 后端集成`easy-workflow`流程引擎，支持或签、并签、判断、会签、抄送、自动化
  - 自定义执行单元可用于绑定工作节点和任务模版，每个执行单元支持绑定多个变量，以满足不同任务场景的灵活配置需求
  - 支持接收企业微信OA回调消息、转换自动化任务执行
    - 接收企业微信数据需自行开发，只需将消息推送至 wechat_callback_events 消息队列。如有需要，可联系我提供具体方案与实现思路。
  - 自动化任务支持变量篡改、输入篡改及任务重试机制，以满足更灵活的任务控制与异常处理需求
  - 集成飞书消息通知、消息回调 `pass` `reject` `progress` `cc` 相应业务处理
  - 定时自动化任务 如：申请权限定时回收
- 用户权限
  - 支持LDAP、账号密码登录方式
  - 支持前端动态菜单、按钮、后端API鉴权
- 排班管理
  - 支持自定义创建多个排班计划
  - 可灵活配置轮换周期、启停时间及多个排班组，以满足多样化的排班需求

## 技术栈
- 数据库：MongoDB、MySQL
- 中间件：Redis Stack、kafka、Etcd
- 框架：Gin、Gorm、Wire、Casbin、Easy-Workflow

## 目录结构
```
.
├── config # 配置文件
│   └── example.yaml
├── deploy # CICD
│   ├── docker-compose.yaml
│   ├── Dockerfile
│   └── prod.yaml
├── docs  # 文档
│   └── img
├── go.mod
├── go.sum
├── internal
│   ├── model      # CMDB - 模型 CI
│   ├── attribute  # CMDB - 字段属性
│   ├── resource   # CMDB - 资产数据
│   ├── relation   # CMDB - 关联关系
│   ├── runner     # 工单 - 执行器
│   ├── task       # 工单 - 自动化任务
│   ├── template   # 工单 - 模版
│   ├── worker     # 工单 - 工作节点
│   └── workflow   # 工单 - 流程绑定
│   ├── codebook   # 工单 - 代码库
│   ├── engine     # 工单 - easyflow engine
│   ├── event      # 工单 - easyflow event 
│   ├── order      # 工单 - 工单信息
│   ├── user       # 权限 - 用户模块
│   ├── role       # 权限 - 角色管理
│   ├── endpoint   # 权限 - API接口管理
│   ├── department # 权限 - 人员分组管理
│   ├── policy     # 权限 - 集成casbin 
│   ├── menu       # 权限 - 菜单信息
│   ├── permission # 权限 - 整合鉴权
│   ├── pkg        # 通用工具包
├── ioc # 依赖注入
│   ├── app.go
│   ├── casbin.go
│   ├── db.go
│   ├── etcd.go
│   ├── gin.go
│   ├── job.go
│   ├── ldap.go
│   ├── mq.go
│   ├── mysql.go
│   ├── redis.go
│   ├── redisearch.go
│   ├── session.go
│   ├── wire_gen.go
│   ├── wire.go
│   └── workwx.go
├── LICENSE
├── main.go # 启动
├── pkg # 二次封装
│   ├── cryptox
│   ├── ginx
│   ├── hash
│   ├── mongox
│   ├── mqx
│   ├── registry
│   └── tools
├── README.md
└── Taskfile.yaml # 类似于makeflie
```

## CMDB
![](docs/img/cmdb.png)

## 工单系统
![](docs/img/order.png)

### 自动化任务-设计流程图
自动化任务在整改系统里面设计比较有趣的地方，也较为复杂。

通过下面这张图，可以更好的理解任务运行流程，以及如何排查定位问题。
![](docs/img/自动化任务-设计流程图.png)

## 权限控制
![](docs/img/permission.png)

## 排班管理
![](docs/img/scheduling.png)
