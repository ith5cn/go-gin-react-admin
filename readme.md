<div align="center">

# Gin React Admin

**基于 Go + React 的全栈后台管理系统**

从账号权限、动态菜单到代码生成，为业务后台提供可扩展的基础工程。

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-149ECA?logo=react&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)
![Ant Design](https://img.shields.io/badge/Ant_Design-6-0170FE?logo=antdesign&logoColor=white)

[快速开始](#快速开始) · [功能一览](#功能一览) · [开发指南](#开发指南) · [反馈问题](https://github.com/ith5cn/go-gin-react-admin/issues) · [NestJS 版本](https://github.com/ith5cn/nestjs-react-admin)

</div>

## 项目介绍

Gin React Admin 是一个前后端分离的后台管理项目，后端使用 Gin + GORM，前端使用 React + TypeScript + Ant Design。项目将用户、角色、菜单、组织、日志等常见后台能力整合到同一套工程中，适合学习 Go 全栈开发，也适合作为 CMS、内部管理平台和业务管理端的开发起点。

项目持续迭代中，重点围绕三个方向建设：

- **完整的后台基础能力**：串联登录认证、角色权限、动态路由和常用管理模块，减少从零搭建的重复工作。
- **面向开发的效率工具**：通过数据库表导入、字段配置和代码预览，生成 Go 后端及 React 前端基础代码。
- **易于理解和扩展的工程结构**：后端按路由、接口、业务、模型分层，前端按页面、组件、接口和状态组织代码。

## 功能一览

| 模块 | 已有能力 |
| --- | --- |
| 登录认证 | JWT 双令牌、Redis 令牌状态校验、单端 / 多端登录 |
| 用户与权限 | 用户、角色、菜单、按钮权限、接口权限校验、用户数据范围控制 |
| 组织管理 | 部门树、岗位管理、用户与角色关联 |
| 动态菜单 | 后端下发菜单与路由，前端按用户权限加载页面 |
| 系统配置 | 配置分组、配置项维护、批量更新 |
| 数据字典 | 字典类型、字典数据、前端字典展示 |
| 附件管理 | 文件上传、附件列表、资源分类与删除 |
| 通知公告 | 公告内容维护与状态管理 |
| 日志审计 | 登录日志、操作日志 |
| 系统监控 | 在线用户管理、服务器运行信息 |
| 定时任务 | Cron 任务配置、执行控制与运行日志 |
| 代码生成 | 数据表导入、字段配置、Go / React 代码预览与生成 |
| 数据维护 | 表结构查看、表优化、碎片清理、软删除数据回收站 |
| 首次安装 | MySQL / Redis 连通性检测、SQL 导入、环境配置与安装锁写入 |

## 技术栈

版本以 [后端依赖](server/go.mod) 和 [前端依赖](web/package.json) 为准。

| 层级 | 技术 |
| --- | --- |
| 后端 | Go 1.25、Gin、GORM、MySQL |
| 认证与缓存 | JWT（golang-jwt/v5）、Redis |
| 日志与任务 | Zap、robfig/cron |
| 前端 | React 19、TypeScript 5.9、Vite 8 |
| UI 与样式 | Ant Design 6、Tailwind CSS 4、Lucide React |
| 路由与状态 | React Router 7、Zustand 5 |
| 网络与国际化 | Axios、i18next |
| 包管理 | Go Modules、pnpm |

## 快速开始

### 1. 准备环境

| 依赖 | 版本 / 要求 |
| --- | --- |
| Go | 1.25 或更高 |
| Node.js | 建议使用 22.12+ 的 Node.js 22；当前 Vite 要求 `^20.19.0 \|\| >=22.12.0` |
| pnpm | 10，与 `web/mise.toml` 保持一致 |
| MySQL | 建议使用 8.4；仓库初始化 SQL 导出自 MySQL 8.4.7 |
| Redis | 可正常连接的 Redis 服务 |

开始安装前，请先启动 MySQL 和 Redis，并准备好连接地址及账号。MySQL 账号需要具备创建目标数据库、建表和导入数据的权限。

> 初始化 SQL 包含 `DROP TABLE IF EXISTS`。请使用专用的空数据库体验项目，避免覆盖已有业务表。

### 2. 获取代码

```bash
git clone https://github.com/ith5cn/go-gin-react-admin.git
cd go-gin-react-admin
```

### 3. 启动后端

在项目根目录打开终端：

```bash
cd server
go mod download
go run .
```

后端默认监听 `http://localhost:8080`。首次启动时，如果没有 `runtime/install.lock`，服务会进入安装模式，无需提前创建 `.env`。

请保持在 `server/` 目录运行后端：环境配置、初始化 SQL 和安装锁均使用相对路径读取。

### 4. 启动前端

在项目根目录另开一个终端：

```bash
cd web
pnpm install
pnpm dev
```

前端默认地址为 `http://localhost:5173`，实际端口以 Vite 终端输出为准。开发服务器会将 `/api` 请求代理至后端，并重写为 `/api/v1`；上传资源通过 `/uploads` 代理访问。

### 5. 完成首次安装

打开 [安装向导](http://localhost:5173/install)，按页面提示操作：

1. 填写 MySQL、Redis 连接信息，并检测连通性。
2. 选择初始化文件 `database/ai_system.sql`。
3. 将默认 JWT 密钥替换为自己的随机密钥，执行安装。
4. 安装完成后进入登录页。

安装向导会创建目标数据库、导入所选 SQL、写入 `server/.env`，并生成 `server/runtime/install.lock`。自定义初始化 SQL 可以放在 `server/sql/` 或 `server/database/` 下，安装页会扫描这两个目录中的 `.sql` 文件。

仓库自带 SQL 的初始管理员账号：

| 账号 | 密码 |
| --- | --- |
| `admin` | `123456` |

首次登录后请修改默认密码。若使用自定义 SQL，账号以对应的初始化数据为准。需要定时任务时，在唯一的调度实例设置 `CRON_ENABLED=true` 并重启，其他 HTTP 实例保持关闭。

## 环境配置

本地运行时，后端从 `server/.env` 加载配置，也支持直接读取系统环境变量。安装向导会生成配置；如需手动维护，可参考 [基础示例](server/.env.example)、[开发环境示例](server/.env.development.example) 和 [生产环境示例](server/.env.production.example)。

常用配置如下，数据库变量前缀为 `AI_SYSTEM_MYSQL_`：

```dotenv
SERVER_ADDR=:8080
ROUTER_PREFIX=/api/v1

AI_SYSTEM_MYSQL_HOST=127.0.0.1
AI_SYSTEM_MYSQL_PORT=3306
AI_SYSTEM_MYSQL_USER=root
AI_SYSTEM_MYSQL_PASSWORD=your_mysql_password
AI_SYSTEM_MYSQL_DB=ai_system
AI_SYSTEM_MYSQL_CONFIG=charset=utf8mb4&parseTime=True&loc=Local

REDIS_MODE=single
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=replace_with_a_long_random_secret
JWT_ACCESS_EXPIRES_MINUTE=120
JWT_REFRESH_EXPIRES_HOUR=168
JWT_LOGIN_MODE=multi
```

- `JWT_LOGIN_MODE` 支持 `single` 和 `multi`，分别对应单端与多端登录。
- Redis 参与登录令牌的状态校验，正常使用系统需要保持 Redis 可用。
- 单独创建 `.env` 不会标记系统为已安装，首次使用仍需完成安装向导。
- 修改监听端口或 API 前缀时，同步调整 [Vite 代理配置](web/vite.config.ts)。前端接口基址由 `VITE_APP_BASE_API` 控制，默认 `/api`。

完整配置项见 [server/config](server/config)。请勿将实际使用的 `.env` 或服务凭据提交到仓库。

## 开发指南

### 目录结构

```text
go-gin-react-admin/
├── server/                   # Go 后端
│   ├── api/                  # 参数绑定、校验与 HTTP 响应
│   ├── config/               # 环境变量与配置读取
│   ├── database/             # 初始化 SQL 与数据库相关文件
│   ├── middleware/           # 认证、权限、日志、CORS 等
│   ├── model/                # 数据模型与请求 / 响应 DTO
│   ├── router/               # 路由注册
│   ├── service/              # 业务逻辑、事务与数据查询
│   ├── setup/                # MySQL、Redis、日志初始化
│   └── utils/                # 通用工具
├── web/                      # React 前端
│   └── src/
│       ├── api/              # 接口请求封装
│       ├── components/       # 通用与业务组件
│       ├── locales/          # 国际化资源
│       ├── pages/            # 页面与业务模块
│       ├── routers/          # 路由配置与访问守卫
│       ├── store/            # Zustand 状态
│       └── utils/            # 通用工具
└── readme.md
```

### 常用命令

| 工作目录 | 命令 | 用途 |
| --- | --- | --- |
| `server/` | `go run .` | 启动后端 |
| `server/` | `go test ./...` | 运行后端测试 |
| `server/` | `go build -o gin-react-admin .` | 构建后端可执行文件 |
| `web/` | `pnpm dev` | 启动前端开发服务器 |
| `web/` | `pnpm lint` | 运行 ESLint |
| `web/` | `pnpm build` | TypeScript 检查并构建至 `dist/` |
| `web/` | `pnpm preview` | 本地预览前端构建产物 |

### 扩展业务模块

后端请求链路为 `router → middleware → api → service → model / database`。新增功能时，路由层负责注册接口，API 层负责参数与响应，业务逻辑和事务放在 service 层。

前端菜单和动态路由由后端用户上下文下发，统一在 Layout 初始化。新增页面时，需要同时配置组件映射、菜单及角色授权；按钮权限码应与后端接口权限保持一致。

对于标准 CRUD 页面，可以从代码生成模块导入数据表，配置字段与展示方式，预览后生成代码。生成结果作为业务开发起点，仍需按实际需求补充校验、权限和测试。

## 构建与部署

前端通过 `pnpm build` 生成静态文件，后端通过 `go build` 生成可执行文件。部署时需配置 SPA 路由回退，并将 `/api` 和 `/uploads` 转发至后端服务。

仓库已提供以下容器化配置：

- 后端：[Dockerfile](server/Dockerfile)、[开发 Compose](server/docker-compose.dev.yml)、[生产 Compose](server/docker-compose.prod.yml) 和 [Makefile](server/Makefile)。
- 前端：[Dockerfile](web/Dockerfile)、[Compose](web/docker-compose.yml) 和 [Nginx 配置](web/nginx.conf)。

这些配置需要按部署环境调整：后端 Compose 仅启动应用服务，MySQL 和 Redis 需单独准备；前端 Dockerfile 当前使用 Node.js 18，构建前需升级至上述环境要求，Nginx 的 API 代理也需配置后启用。

后端运行目录需保留 `.env`、`runtime/install.lock` 与 `runtime/uploads/`。容器部署时，Compose 已挂载 `runtime` 数据卷，后续启动的环境变量应同步写入对应的 Compose 环境文件。

## 常见问题

**为什么启动后进入安装页？**

系统通过 `server/runtime/install.lock` 判断安装状态。首次运行需要完成安装；已有安装仍出现此情况时，请检查后端工作目录和 `runtime` 是否正确保留。

**为什么登录后又回到登录页？**

认证同时校验 JWT 和 Redis 中的令牌状态。请检查 Redis 连接、JWT 密钥是否变更，以及令牌是否过期或被撤销。

**为什么前端接口请求失败？**

先确认后端已启动，再检查前端接口基址、Vite 代理目标和 API 前缀是否一致。生产环境还需配置反向代理，Vite 的开发代理不会包含在静态构建产物中。

## 参与贡献

欢迎通过 [Issue](https://github.com/ith5cn/go-gin-react-admin/issues) 反馈问题或讨论新功能，也欢迎提交 Pull Request。

- 反馈问题时，请附上运行环境、复现步骤、预期结果与相关日志，并移除敏感信息。
- 提交代码前，请运行与改动相关的测试；前端改动需通过 `pnpm lint` 和 `pnpm build`。
- 提交信息使用 `feat`、`fix`、`docs`、`refactor` 等 Conventional Commits 前缀，一个 PR 聚焦一个明确目的。

## 后续计划

- [ ] 完善代码生成模板与生成后业务扩展文档。
- [ ] 补充关键业务的单元测试和端到端测试。
- [ ] 完善容器部署配置与部署文档。
- [ ] 补充项目截图、操作演示和使用文档。

## 相关项目

如果你更熟悉 Node.js / TypeScript，可以参考 [NestJS React Admin](https://github.com/ith5cn/nestjs-react-admin)。两个项目围绕相近的后台管理需求，分别探索 Go 与 NestJS 技术栈下的实现。

## 许可证

当前仓库尚未包含 `LICENSE` 文件，许可证待确定并补充。

## 多项目模板升级

模板版本在 `server/VERSION`。已安装项目升级前，在 `server/` 执行 `go run ./cmd/migrate status` 审阅待执行版本，备份后执行 `go run ./cmd/migrate up`。HTTP 启动只检查版本，不再执行 DDL；首次安装由向导执行迁移。不要向已有数据库重新导入全量初始化 SQL。

- [更新记录](CHANGELOG.md)
- [数据库迁移、任务部署与版本发布](docs/template-upgrade.md)
- [PHP 兼容示例、模块依赖与生成代码边界](docs/php-migration.md)

CI 阻断 Go 测试/构建及前端类型检查/构建失败；现有前端 ESLint 问题单独报告，等待后续清理。
