# v1.9.5 增量数据迁移记录

## 核心变更

在多租户改造完成后，我们将历史没有租户（即不存在 `tenant_id` 字段）或租户 ID 为零的数据批量归属、迁移到指定的租户空间：**租户 2**。

### 影响的集合

迁移操作会全自动在以下 8 个 CMDB 核心集合上执行，并将这些历史数据完全洗白到租户 2 下：

1. `c_resources` (资产资源)
2. `c_model` (模型)
3. `c_model_group` (模型分组)
4. `c_attribute` (模型属性)
5. `c_attribute_group` (属性分组)
6. `c_relation_type` (关系类型)
7. `c_relation_model` (模型关系)
8. `c_relation_resource` (资产关联)

---

## 升级前置操作 (Before)

* **全自动数据备份**：升级脚本运行的 `Before` 阶段会自动调用 `BackupManager`，依次对这 8 个集合进行物理 JSON 备份并存储。如果迁移过程中发生任何不可预测的错误，可以通过备份一键进行安全恢复。

---

## 提交变更 (Commit)

* **数据清洗订正**：
  * 通过强力高效的 `UpdateMany` 批量执行器，对上述 8 个集合运行订正：
    ```go
    bson.M{"$or": []bson.M{
        {"tenant_id": bson.M{"$exists": false}},
        {"tenant_id": 0},
    }}
    ```
  * 将以上匹配出的所有脏历史数据，一律安全、稳健地重写为 `tenant_id: 2`。
