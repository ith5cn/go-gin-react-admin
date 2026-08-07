---
description: 基于仓库历史的提交、分支和交付规则
---

# Git 工作流

- 提交采用仓库已在使用的 Conventional Commits：`feat`、`fix`、`refactor`、`chore`、`docs`、`test`，可带 `auth`、`system`、`codegen`、`backend` 等 scope。
- subject 简洁描述一个可审查目的，可使用中文；示例：`fix(auth): 修复刷新令牌并发重试`。
- 一个提交只包含一个逻辑变更；提交前查看 diff 和状态，避免混入 `.env`、构建产物、缓存或用户的无关改动。
- 分支建议使用 `feat/<短描述>`、`fix/<短描述>`、`chore/<短描述>`。
- 不使用 `--no-verify` 绕过钩子；提交前执行与变更匹配的测试、构建和 lint。
- schema、权限码或生成器契约变更需将代码、`ai_system.sql`、模板、测试和文档作为同一交付范围同步检查。
