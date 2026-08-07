# gin-react-admin

Go + React 全栈后台管理系统，包含认证权限、系统管理、首次安装、代码生成和数据维护能力。

## 技术栈

- 后端：Go 1.25、Gin 1.12、GORM 1.31（MySQL）、Redis、JWT、Zap
- 前端：React 19、TypeScript 5.9、Vite 8、Ant Design 6、Zustand 5、React Router 7、Tailwind CSS 4
- 包管理：后端 `go mod`，前端 `pnpm`

## 常用命令

### 后端（`server/`）

- 运行：`go run main.go`
- 测试：`go test ./...`
- 格式化：`gofmt -w <files>`
- 依赖整理：`go mod tidy`

后端从当前工作目录的 `.env` 读取配置，也支持直接使用系统环境变量。首次启动若无安装锁，会启用 `/install` 向导并跳过业务数据库、Redis和定时任务初始化。

### 前端（`web/`）

- 安装依赖：`pnpm install`
- 开发运行：`pnpm dev`
- 构建：`pnpm build`（`tsc -b && vite build`）
- Lint：`pnpm lint`
- 预览：`pnpm preview`
- mise 流程：`mise trust`、`mise install`、`mise run install`、`mise run dev`

Vite 将 `/api` 重写为后端 `/api/v1`，并代理 `/uploads` 到 `http://localhost:8080`。

## 目录结构

```text
gin-react-admin/
├── server/
│   ├── api/                 # Gin handler：绑参数、调 service、组织响应
│   ├── service/             # 业务逻辑、定时任务、代码生成
│   ├── model/               # GORM model、request/response DTO
│   ├── router/              # 公共/私有路由和生成路由注册
│   ├── middleware/          # JWT、权限、日志、Recovery、CORS
│   ├── config/              # 环境变量配置
│   ├── setup/               # MySQL、Redis、Zap 初始化
│   └── database/            # ai_system 全量初始化 SQL
└── web/
    ├── src/api/             # request 封装的接口函数
    ├── src/store/           # Zustand 状态
    ├── src/routers/         # 静态/动态路由与 AuthGuard
    ├── src/pages/           # 页面组件
    └── src/components/      # 通用与布局组件
```

## 关键环境变量

`SERVER_ADDR`、`ROUTER_PREFIX`、`DB_AI_SYSTEM_HOST`、`DB_AI_SYSTEM_PORT`、`DB_AI_SYSTEM_USER`、`DB_AI_SYSTEM_PASSWORD`、`DB_AI_SYSTEM_DBNAME`、`REDIS_ADDR`、`JWT_SECRET`、`JWT_ACCESS_EXPIRES_MINUTE`、`JWT_REFRESH_EXPIRES_HOUR`、`JWT_LOGIN_MODE`。

环境变量的默认值和完整清单以 `server/config/go_*.go` 为准。仓库当前没有 `.env.example`，不要提交真实 `.env`。

## 项目现状与约束

- 后端已有 7 个 `_test.go` 文件，主要覆盖路由、操作日志、定时任务和代码生成；认证、权限、数据库集成仍缺少系统测试。
- 前端尚未配置测试框架，也没有前端测试文件；`pnpm build` 和 `pnpm lint` 是最低验证要求。
- 动态菜单由 `/system/user` 下发，经规范化后生成路由；`Layout` 统一初始化用户上下文，登录页不要重复调用 `initUserContext`。
- access token 同时校验 JWT 和 Redis 中的 jti；Redis 不可用、token 被撤销或会话失效时，认证接口会拒绝访问。
- 数据库没有独立 migration 工具；schema 和种子数据以 `server/database/ai_system.sql` 为当前基准。
- `api/system/common.go` 的导出包装和生成目录是代码生成器稳定契约，修改时必须同步模板与测试。

## 规则

全程适用：

- `.codex/rules/coding-style.md`
- `.codex/rules/testing.md`
- `.codex/rules/security.md`
- `.codex/rules/git-workflow.md`

条件规则：

- 前端：`.codex/rules/frontend.md`（`web/**`）
- 后端：`.codex/rules/backend-api.md`（`server/**`）
- 数据库：`.codex/rules/database.md`（`server/database/**`、`server/model/**`）
