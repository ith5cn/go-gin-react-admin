---
description: MySQL、GORM model、schema 与查询变更规则
paths:
  - server/database/**
  - server/model/**
  - server/service/**
  - server/setup/gorm/**
---

# 数据库规则

- 当前无独立 migration 工具；`server/database/ai_system.sql` 是全量 DDL 与种子数据基准，schema、索引、权限菜单变化需同步更新。
- model 放在 `server/model/system/` 或对应模块，使用显式 `TableName()`、`column` tag，并遵循现有 camelCase JSON 字段约定。
- 不假设 model 嵌入 `gorm.Model`；本项目多使用 `create_time`、`update_time`、`delete_time` 等显式字段，按实际表结构建模。
- model 不承载业务流程；请求与响应 DTO 放在 `model/.../request`、`model/.../response`。
- 查询值使用 GORM 参数绑定；动态标识符通过白名单映射。分页响应使用项目 `response.PageResult` 的 `list` 和 `total`。
- 软删除逻辑遵循各表实际 `delete_time` 设计；没有软删列的表不得强行追加过滤，代码生成器也必须正确区分软删与硬删。
- 多连接通过 `gormInit.Gorm.Get(<name>)` 获取；连接名称、DSN 与池参数统一由 `config/` 和 `setup/gorm/` 管理。
- 变更表结构时检查初始化向导、代码生成、数据维护页面和已有数据库升级路径；在没有 migration 工具时明确提供兼容性 SQL 或升级说明。
