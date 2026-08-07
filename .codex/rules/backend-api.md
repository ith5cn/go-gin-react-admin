---
description: Gin 后端的分层、路由、错误、鉴权与代码生成契约
paths: server/**
---

# 后端 API 规则

- 保持 `router -> api -> service -> model/setup` 分层：router 注册中间件和路径，api 绑定参数并响应，service 实现业务，model 表达持久化结构与 DTO。
- 公共路由挂 `PublicGroup`，业务路由挂带 `JWTAuth` 和操作日志的 `PrivateGroup`；需要细粒度授权时再挂 `middleware.Perm`。
- 默认业务前缀是 `ROUTER_PREFIX=/api/v1`，后端同时兼容 `/api`；修改路由时验证两套注册不会引入冲突或重复副作用。
- 请求优先使用类型化 `bindJSON[T]`；响应复用 `response.Success`、`response.Fail`、`response.FailWithHTTP` 和 `successOrFail`。
- service 返回业务错误或内部错误，不直接依赖 Gin。api 用 `errors.Is`/`errors.As` 分类；内部错误写 Zap 日志并返回泛化 SystemError。
- JWT context 中的 `user_id`、`username` 是身份来源；权限码格式沿用现有 `system/<module>/<action>`，并与菜单按钮记录一致。
- 行级数据权限复用 `UserDataScope` 和既有查询注入逻辑，禁止信任客户端传入的操作人 ID 或自行复制部门过滤规则。
- 定时任务创建/更新须校验表达式，写操作后重载 scheduler；任务函数处理自身错误，禁止 panic 越过调度边界。
- `api/system/common.go` 的 `QueryMap`、`BindJSONMap`、`SuccessOrFail` 及 `generated/` 结构属于代码生成契约；修改时同步 `codegen_templates.go` 与相关测试。
- 首次安装状态下业务基础设施未初始化；新增全局中间件或路由不能假设数据库和 Redis 一定可用。
