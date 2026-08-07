---
description: Go、TypeScript 与 React 的项目代码风格
---

# 代码风格

## Go

- 使用 `gofmt`，按 Go 惯例命名：导出标识符 PascalCase，包内标识符 camelCase；缩写保持一致。
- import 使用标准库、项目包、第三方包的分组；已有别名如 `systemService`、`gormInit` 保持目录语义。
- 普通错误沿调用链返回；启动阶段不可恢复错误才记录 Fatal。业务错误使用项目的 `BizError`/sentinel error 和 `errors.Is`/`errors.As`，禁止依赖错误字符串分支。
- 使用项目 Zap logger，禁止用 `fmt.Print*` 作为运行日志。
- handler 保持薄层，业务逻辑放 `service/`；model 只表达持久化结构和 DTO。
- 导出类型和函数写有意义的 Go doc；其他注释解释约束和原因，不复述代码。

## TypeScript / React

- 遵循仓库现有风格：2 空格、单引号、无分号，并通过 ESLint 与 TypeScript strict 校验。
- 组件和类型使用 PascalCase，函数和变量使用 camelCase，hook 以 `use` 开头；使用函数组件和 hooks。
- 优先 `type` 与明确泛型，避免新增 `any`；第三方边界暂时无法收窄时，把转换限制在边界层并说明原因。
- 优先 `@/` 别名导入项目模块；同一文件内保持外部依赖、内部别名、相对模块的清晰分组。
- 不引入未使用变量、参数或有副作用但未声明目的的 import；`tsconfig.app.json` 已开启相应严格检查。
